package httptool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tempiex/pi/internal/tool"
)

type Input struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

type Output struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func New() tool.AnyTool {
	return &tool.Tool[Input, Output]{
		Name:          "http",
		Description:   "Make an HTTP request",
		ExecutionMode: "sync",
		Execute: func(ctx context.Context, input Input) (Output, error) {
			timeout := time.Duration(input.Timeout) * time.Second
			if timeout == 0 {
				timeout = 30 * time.Second
			}
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			method := input.Method
			if method == "" {
				method = http.MethodGet
			}

			var body io.Reader
			if input.Body != "" {
				body = strings.NewReader(input.Body)
			}

			req, err := http.NewRequestWithContext(ctx, method, input.URL, body)
			if err != nil {
				return Output{}, fmt.Errorf("create request: %w", err)
			}
			for k, v := range input.Headers {
				req.Header.Set(k, v)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return Output{}, err
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return Output{}, err
			}

			headers := make(map[string]string, len(resp.Header))
			for k := range resp.Header {
				headers[k] = resp.Header.Get(k)
			}

			return Output{
				StatusCode: resp.StatusCode,
				Headers:    headers,
				Body:       string(respBody),
			}, nil
		},
	}
}
