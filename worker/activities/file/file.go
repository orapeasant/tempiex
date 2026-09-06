// Package file provides built-in file read and write activities.
package file

import (
	"fmt"
	"os"

	"github.com/tempiex/worker/activity"
)

type ReadInput struct {
	Path string `json:"path"`
}

type ReadOutput struct {
	Content string `json:"content"`
}

func NewReadActivity() activity.AnyActivity {
	return &activity.Activity[ReadInput, ReadOutput]{
		Name:          "file_read",
		Description:   "Read a file from disk",
		ExecutionMode: "parallel",
		Execute: func(_ activity.Context, input ReadInput) (ReadOutput, error) {
			data, err := os.ReadFile(input.Path)
			if err != nil {
				return ReadOutput{}, fmt.Errorf("read file: %w", err)
			}
			return ReadOutput{Content: string(data)}, nil
		},
	}
}

type WriteInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WriteOutput struct {
	BytesWritten int `json:"bytesWritten"`
}

func NewWriteActivity() activity.AnyActivity {
	return &activity.Activity[WriteInput, WriteOutput]{
		Name:          "file_write",
		Description:   "Write content to a file on disk",
		ExecutionMode: "sequential",
		Execute: func(_ activity.Context, input WriteInput) (WriteOutput, error) {
			if err := os.WriteFile(input.Path, []byte(input.Content), 0o644); err != nil {
				return WriteOutput{}, fmt.Errorf("write file: %w", err)
			}
			return WriteOutput{BytesWritten: len(input.Content)}, nil
		},
	}
}
