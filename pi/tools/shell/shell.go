package shell

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/tempiex/pi/internal/tool"
)

type Input struct {
	Cmd     string   `json:"cmd"`
	Args    []string `json:"args,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

type Output struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

func New() tool.AnyTool {
	return &tool.Tool[Input, Output]{
		Name:          "shell",
		Description:   "Execute a shell command",
		ExecutionMode: "sync",
		Execute: func(ctx context.Context, input Input) (Output, error) {
			timeout := time.Duration(input.Timeout) * time.Second
			if timeout == 0 {
				timeout = 30 * time.Second
			}
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, input.Cmd, input.Args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
					err = nil
				}
			}
			return Output{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: exitCode,
			}, err
		},
	}
}
