// Package activity is the typed activity runtime for a Tempiex worker.
//
// An activity is a plain Go function with typed input and output, registered
// under a name that matches the scheduled ActivityType. Executor implements
// worker.ActivityHandler: it decodes the task payload, runs the activity, and
// reports the outcome back to Tempiex.
//
// Nothing here knows about AI agents. An agent that runs as an activity is a
// separate handler (see the agent bridge in docs/spec/08-worker.md).
package activity

import "encoding/json"

// Activity is a typed activity implementation.
type Activity[I, O any] struct {
	Name        string
	Description string
	// Execute performs the activity. A returned error fails the activity task.
	Execute func(ctx Context, input I) (O, error)
	// BeforeExec may validate or reject the input. A returned error fails the
	// task without calling Execute.
	BeforeExec func(ctx Context, input I) error
	// AfterExec may transform the output, or convert an Execute error into a
	// successful result (and vice versa), before the outcome is reported.
	AfterExec func(ctx Context, input I, output O, err error) (O, error)
	// ExecutionMode is advisory metadata surfaced on /api/activities.
	ExecutionMode string
}

// Descriptor is the model-independent description of a registered activity.
type Descriptor struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ExecutionMode string `json:"executionMode"`
}

// AnyActivity is the type-erased form stored in a Registry.
type AnyActivity interface {
	Descriptor() Descriptor
	Run(ctx Context, inputJSON []byte) ([]byte, error)
}

func (a *Activity[I, O]) Descriptor() Descriptor {
	return Descriptor{
		Name:          a.Name,
		Description:   a.Description,
		ExecutionMode: a.ExecutionMode,
	}
}

// Run decodes inputJSON, runs the Before/Execute/After chain, and encodes the
// output. Decode failures are returned as InputError so the caller can report
// them distinctly from an activity that simply failed.
func (a *Activity[I, O]) Run(ctx Context, inputJSON []byte) ([]byte, error) {
	var input I
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		return nil, &InputError{Err: err}
	}

	if a.BeforeExec != nil {
		if err := a.BeforeExec(ctx, input); err != nil {
			return nil, err
		}
	}

	output, execErr := a.Execute(ctx, input)

	if a.AfterExec != nil {
		var err error
		output, err = a.AfterExec(ctx, input, output, execErr)
		if err != nil {
			return nil, err
		}
	} else if execErr != nil {
		return nil, execErr
	}

	return json.Marshal(output)
}
