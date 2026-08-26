package tool

import (
	"context"
	"encoding/json"
	"testing"
)

type testInput struct {
	Value string `json:"value"`
}
type testOutput struct {
	Echo string `json:"echo"`
}

func newTestTool() AnyTool {
	return &Tool[testInput, testOutput]{
		Name:          "test-tool",
		Description:   "echoes input",
		ExecutionMode: "sync",
		Execute: func(_ context.Context, in testInput) (testOutput, error) {
			return testOutput{Echo: in.Value}, nil
		},
	}
}

func TestRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(newTestTool())

	got, ok := r.Get("test-tool")
	if !ok {
		t.Fatal("expected tool to be found")
	}
	if got.Info().Name != "test-tool" {
		t.Fatalf("unexpected name: %s", got.Info().Name)
	}
}

func TestList(t *testing.T) {
	r := NewRegistry()
	r.Register(newTestTool())
	list := r.List()
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}
}

func TestRun(t *testing.T) {
	r := NewRegistry()
	r.Register(newTestTool())
	tool, _ := r.Get("test-tool")

	input, _ := json.Marshal(testInput{Value: "hello"})
	out, err := tool.Run(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	var o testOutput
	_ = json.Unmarshal(out, &o)
	if o.Echo != "hello" {
		t.Fatalf("expected hello, got %s", o.Echo)
	}
}
