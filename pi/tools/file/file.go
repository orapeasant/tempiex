package file

import (
	"context"
	"fmt"
	"os"

	"github.com/tempiex/pi/internal/tool"
)

type ReadInput struct {
	Path string `json:"path"`
}

type ReadOutput struct {
	Content string `json:"content"`
}

func NewReadTool() tool.AnyTool {
	return &tool.Tool[ReadInput, ReadOutput]{
		Name:          "file_read",
		Description:   "Read a file from disk",
		ExecutionMode: "sync",
		Execute: func(_ context.Context, input ReadInput) (ReadOutput, error) {
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

func NewWriteTool() tool.AnyTool {
	return &tool.Tool[WriteInput, WriteOutput]{
		Name:          "file_write",
		Description:   "Write content to a file on disk",
		ExecutionMode: "sync",
		Execute: func(_ context.Context, input WriteInput) (WriteOutput, error) {
			if err := os.WriteFile(input.Path, []byte(input.Content), 0644); err != nil {
				return WriteOutput{}, fmt.Errorf("write file: %w", err)
			}
			return WriteOutput{BytesWritten: len(input.Content)}, nil
		},
	}
}
