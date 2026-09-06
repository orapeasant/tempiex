// Package httptool provides a built-in activity that makes an HTTP request.
package httptool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tempiex/worker/activity"
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

func New() activity.AnyActivity {
	return &activity.Activity[Input, Output]{
		Name:          "http",
		Description:   "Make an HTTP request",
		ExecutionMode: "parallel",
		Execute: func(ctx activity.Context, input Input) (Output, error) {
			timeout := time.Duration(input.Timeout) * time.Second
			if timeout == 0 {
				timeout = 30 * time.Second
			}
			reqCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			method := input.Method
			if method == "" {
				method = http.MethodGet
			}

			var body io.Reader
			if input.Body != "" {
				body = strings.NewReader(input.Body)
			}

			req, err := http.NewRequestWithContext(reqCtx, method, input.URL, body)
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
