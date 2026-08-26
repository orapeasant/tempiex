package tool

import (
	"context"
	"encoding/json"
)

type Tool[I, O any] struct {
	Name          string
	Description   string
	Execute       func(ctx context.Context, input I) (O, error)
	BeforeExec    func(ctx context.Context, input I) error
	AfterExec     func(ctx context.Context, input I, output O, err error) (O, error)
	ExecutionMode string
}

type ToolInfo struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ExecutionMode string `json:"executionMode"`
}

type AnyTool interface {
	Info() ToolInfo
	Run(ctx context.Context, inputJSON []byte) ([]byte, error)
	RunBefore(ctx context.Context, inputJSON []byte) error
}

func (t *Tool[I, O]) Info() ToolInfo {
	return ToolInfo{
		Name:          t.Name,
		Description:   t.Description,
		ExecutionMode: t.ExecutionMode,
	}
}

func (t *Tool[I, O]) RunBefore(ctx context.Context, inputJSON []byte) error {
	if t.BeforeExec == nil {
		return nil
	}
	var input I
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		return err
	}
	return t.BeforeExec(ctx, input)
}

func (t *Tool[I, O]) Run(ctx context.Context, inputJSON []byte) ([]byte, error) {
	var input I
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		return nil, err
	}

	if t.BeforeExec != nil {
		if err := t.BeforeExec(ctx, input); err != nil {
			return nil, err
		}
	}

	output, execErr := t.Execute(ctx, input)

	if t.AfterExec != nil {
		var err error
		output, err = t.AfterExec(ctx, input, output, execErr)
		if err != nil {
			return nil, err
		}
	} else if execErr != nil {
		return nil, execErr
	}

	return json.Marshal(output)
}
