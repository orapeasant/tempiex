package activity

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

func echoActivity() Activity[testInput, testOutput] {
	return Activity[testInput, testOutput]{
		Name:          "test-activity",
		Description:   "echoes input",
		ExecutionMode: "parallel",
		Execute: func(_ Context, in testInput) (testOutput, error) {
			return testOutput{Echo: in.Value}, nil
		},
	}
}

// nullContext is the minimal Context an activity needs when run outside a
// dequeued task.
type nullContext struct{ context.Context }

func (nullContext) Info() Info          { return Info{} }
func (nullContext) Heartbeat(any) error { return nil }
func (nullContext) Publish(any)         {}

func TestRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	Register(r, echoActivity())

	got, ok := r.Get("test-activity")
	if !ok {
		t.Fatal("expected activity to be found")
	}
	if got.Descriptor().Name != "test-activity" {
		t.Fatalf("unexpected name: %s", got.Descriptor().Name)
	}
}

func TestDuplicateRegistrationPanics(t *testing.T) {
	r := NewRegistry()
	Register(r, echoActivity())

	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic on duplicate registration")
		}
	}()
	Register(r, echoActivity())
}

func TestListIsSorted(t *testing.T) {
	r := NewRegistry()
	Register(r, echoActivity())
	Register(r, Activity[testInput, testOutput]{
		Name:    "another",
		Execute: func(_ Context, in testInput) (testOutput, error) { return testOutput{}, nil },
	})

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
	if list[0].Name != "another" || list[1].Name != "test-activity" {
		t.Fatalf("expected name order, got %v", list)
	}
}

func TestRunEncodesOutput(t *testing.T) {
	r := NewRegistry()
	Register(r, echoActivity())
	act, _ := r.Get("test-activity")

	input, _ := json.Marshal(testInput{Value: "hello"})
	out, err := act.Run(nullContext{context.Background()}, input)
	if err != nil {
		t.Fatal(err)
	}

	var o testOutput
	_ = json.Unmarshal(out, &o)
	if o.Echo != "hello" {
		t.Fatalf("expected hello, got %s", o.Echo)
	}
}

func TestRunReportsInputDecodeFailure(t *testing.T) {
	r := NewRegistry()
	Register(r, echoActivity())
	act, _ := r.Get("test-activity")

	_, err := act.Run(nullContext{context.Background()}, []byte("not json"))
	var inputErr *InputError
	if !asInputError(err, &inputErr) {
		t.Fatalf("expected an InputError, got %v", err)
	}
}

func asInputError(err error, target **InputError) bool {
	e, ok := err.(*InputError)
	if ok {
		*target = e
	}
	return ok
}

func TestHooksRunInOrder(t *testing.T) {
	var order []string
	r := NewRegistry()
	Register(r, Activity[testInput, testOutput]{
		Name: "hooked",
		BeforeExec: func(_ Context, _ testInput) error {
			order = append(order, "before")
			return nil
		},
		Execute: func(_ Context, in testInput) (testOutput, error) {
			order = append(order, "execute")
			return testOutput{Echo: in.Value}, nil
		},
		AfterExec: func(_ Context, _ testInput, out testOutput, err error) (testOutput, error) {
			order = append(order, "after")
			out.Echo += "!"
			return out, err
		},
	})

	act, _ := r.Get("hooked")
	input, _ := json.Marshal(testInput{Value: "hi"})
	out, err := act.Run(nullContext{context.Background()}, input)
	if err != nil {
		t.Fatal(err)
	}

	var o testOutput
	_ = json.Unmarshal(out, &o)
	if o.Echo != "hi!" {
		t.Fatalf("expected AfterExec to transform output, got %s", o.Echo)
	}
	if len(order) != 3 || order[0] != "before" || order[1] != "execute" || order[2] != "after" {
		t.Fatalf("unexpected hook order: %v", order)
	}
}
