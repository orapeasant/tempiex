package metrics

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tempiex/pi/internal/config"
)

type Metrics struct {
	ActivityExecutions metric.Int64Counter
	ActivityDuration   metric.Float64Histogram
	SessionsActive     metric.Int64UpDownCounter
	WorkerPollLatency  metric.Float64Histogram
	provider           *sdkmetric.MeterProvider
	tracerProvider     *sdktrace.TracerProvider
}

func Setup(ctx context.Context, cfg config.MetricsConfig, log zerolog.Logger) (*Metrics, error) {
	if cfg.OTel.Endpoint == "" {
		return noopMetrics(), nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName("pi")),
	)
	if err != nil {
		log.Warn().Err(err).Msg("otel resource setup failed; using noop")
		return noopMetrics(), nil
	}

	dialOpts := []grpc.DialOption{}
	if cfg.OTel.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	metricExp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OTel.Endpoint),
		otlpmetricgrpc.WithDialOption(dialOpts...),
	)
	if err != nil {
		log.Warn().Err(err).Msg("otel metric exporter setup failed; using noop")
		return noopMetrics(), nil
	}

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTel.Endpoint),
		otlptracegrpc.WithDialOption(dialOpts...),
	)
	if err != nil {
		log.Warn().Err(err).Msg("otel trace exporter setup failed; using noop")
		_ = metricExp.Shutdown(ctx)
		return noopMetrics(), nil
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
	)

	m := &Metrics{provider: mp, tracerProvider: tp}
	if err := m.initInstruments(mp.Meter("pi")); err != nil {
		return nil, fmt.Errorf("init instruments: %w", err)
	}
	return m, nil
}

func (m *Metrics) initInstruments(meter metric.Meter) error {
	var err error
	m.ActivityExecutions, err = meter.Int64Counter("pi.activity.executions")
	if err != nil {
		return err
	}
	m.ActivityDuration, err = meter.Float64Histogram("pi.activity.duration_ms")
	if err != nil {
		return err
	}
	m.SessionsActive, err = meter.Int64UpDownCounter("pi.sessions.active")
	if err != nil {
		return err
	}
	m.WorkerPollLatency, err = meter.Float64Histogram("pi.worker.poll_latency_ms")
	if err != nil {
		return err
	}
	return nil
}

func (m *Metrics) Tracer() trace.Tracer {
	if m.tracerProvider == nil {
		return tracenoop.NewTracerProvider().Tracer("pi")
	}
	return m.tracerProvider.Tracer("pi")
}

func (m *Metrics) Shutdown(ctx context.Context) error {
	if m.provider != nil {
		_ = m.provider.Shutdown(ctx)
	}
	if m.tracerProvider != nil {
		_ = m.tracerProvider.Shutdown(ctx)
	}
	return nil
}

func noopMetrics() *Metrics {
	nm := noop.NewMeterProvider().Meter("pi")
	m := &Metrics{}
	_ = m.initInstruments(nm)
	return m
}
