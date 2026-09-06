// Package shell provides a built-in activity that runs a shell command.
package shell

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/tempiex/worker/activity"
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

func New() activity.AnyActivity {
	return &activity.Activity[Input, Output]{
		Name:          "shell",
		Description:   "Execute a shell command",
		ExecutionMode: "sequential",
		Execute: func(ctx activity.Context, input Input) (Output, error) {
			timeout := time.Duration(input.Timeout) * time.Second
			if timeout == 0 {
				timeout = 30 * time.Second
			}
			cmdCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, input.Cmd, input.Args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			exitCode := 0
			if err != nil {
				// A non-zero exit is a result, not an activity failure: the
				// workflow decides what a failing command means.
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
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
