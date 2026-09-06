// Package metrics wires the worker's OpenTelemetry instruments and exporters.
//
// When no OTLP endpoint is configured (or exporter setup fails) the worker
// falls back to no-op instruments rather than failing to start: telemetry
// being unavailable must never stop activities from running.
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
)

const serviceName = "worker"

// Metrics holds the worker's instruments plus the providers backing them.
type Metrics struct {
	ActivityExecutions metric.Int64Counter
	ActivityDuration   metric.Float64Histogram
	RunsActive         metric.Int64UpDownCounter
	EventsDropped      metric.Int64Counter

	provider       *sdkmetric.MeterProvider
	tracerProvider *sdktrace.TracerProvider
}

// Setup builds OTLP-backed metrics and tracing. An empty endpoint, or any
// exporter failure, yields working no-op instruments and a nil error.
func Setup(ctx context.Context, endpoint string, insecureConn bool, log zerolog.Logger) (*Metrics, error) {
	if endpoint == "" {
		return Noop(), nil
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	if err != nil {
		log.Warn().Err(err).Msg("otel resource setup failed; using noop")
		return Noop(), nil
	}

	var dialOpts []grpc.DialOption
	if insecureConn {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	metricExp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithDialOption(dialOpts...),
	)
	if err != nil {
		log.Warn().Err(err).Msg("otel metric exporter setup failed; using noop")
		return Noop(), nil
	}

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(dialOpts...),
	)
	if err != nil {
		log.Warn().Err(err).Msg("otel trace exporter setup failed; using noop")
		_ = metricExp.Shutdown(ctx)
		return Noop(), nil
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
	if err := m.initInstruments(mp.Meter(serviceName)); err != nil {
		return nil, fmt.Errorf("init instruments: %w", err)
	}
	return m, nil
}

// Noop returns Metrics whose instruments discard everything. Useful in tests
// and whenever telemetry is disabled.
func Noop() *Metrics {
	m := &Metrics{}
	_ = m.initInstruments(noop.NewMeterProvider().Meter(serviceName))
	return m
}

func (m *Metrics) initInstruments(meter metric.Meter) error {
	var err error
	if m.ActivityExecutions, err = meter.Int64Counter("worker.activity.executions.total"); err != nil {
		return err
	}
	if m.ActivityDuration, err = meter.Float64Histogram("worker.activity.duration"); err != nil {
		return err
	}
	if m.RunsActive, err = meter.Int64UpDownCounter("worker.runs.active"); err != nil {
		return err
	}
	if m.EventsDropped, err = meter.Int64Counter("worker.events.dropped.total"); err != nil {
		return err
	}
	return nil
}

func (m *Metrics) Tracer() trace.Tracer {
	if m.tracerProvider == nil {
		return tracenoop.NewTracerProvider().Tracer(serviceName)
	}
	return m.tracerProvider.Tracer(serviceName)
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
