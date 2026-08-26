package matching

import (
	"context"
	"sync"
	"time"

	"github.com/tempiex/tempiex/internal/persistence"
)

// WorkflowTask represents a workflow task for matching.
type WorkflowTask struct {
	ID               int64
	NamespaceID      string
	WorkflowID       string
	RunID            string
	TaskQueue        string
	ScheduledEventID int64
}

// ActivityTask represents an activity task for matching.
type ActivityTask struct {
	ID               int64
	NamespaceID      string
	WorkflowID       string
	RunID            string
	TaskQueue        string
	ActivityID       string
	ActivityType     string
	ScheduledEventID int64
	Input            []byte
}

type taskQueue struct {
	wfCh  chan *WorkflowTask
	actCh chan *ActivityTask
}

// Engine is the in-memory matching engine with DB fallback.
type Engine struct {
	mu     sync.Mutex
	queues map[string]*taskQueue
	store  persistence.Store
}

// NewEngine creates a new matching engine.
func NewEngine(store persistence.Store) *Engine {
	return &Engine{
		queues: make(map[string]*taskQueue),
		store:  store,
	}
}

func (e *Engine) getOrCreateQueue(key string) *taskQueue {
	e.mu.Lock()
	defer e.mu.Unlock()
	q, ok := e.queues[key]
	if !ok {
		q = &taskQueue{
			wfCh:  make(chan *WorkflowTask, 256),
			actCh: make(chan *ActivityTask, 256),
		}
		e.queues[key] = q
	}
	return q
}

func queueKey(namespaceID, taskQueue string) string {
	return namespaceID + ":" + taskQueue
}

// AddWorkflowTask adds a workflow task to the matching engine.
func (e *Engine) AddWorkflowTask(ctx context.Context, task *WorkflowTask) error {
	q := e.getOrCreateQueue(queueKey(task.NamespaceID, task.TaskQueue))
	select {
	case q.wfCh <- task:
		return nil
	default:
		// Channel full, persist to DB
		return e.store.CreateWorkflowTask(ctx, &persistence.WorkflowTask{
			NamespaceID:      task.NamespaceID,
			WorkflowID:       task.WorkflowID,
			RunID:            task.RunID,
			TaskQueue:        task.TaskQueue,
			ScheduledEventID: task.ScheduledEventID,
		})
	}
}

// PollWorkflowTask polls for a workflow task with a timeout.
func (e *Engine) PollWorkflowTask(ctx context.Context, namespaceID, taskQueue string, timeout time.Duration) (*WorkflowTask, error) {
	// First check DB for buffered tasks
	dbTask, err := e.store.ClaimWorkflowTask(ctx, namespaceID, taskQueue)
	if err != nil {
		return nil, err
	}
	if dbTask != nil {
		return &WorkflowTask{
			ID:               dbTask.ID,
			NamespaceID:      dbTask.NamespaceID,
			WorkflowID:       dbTask.WorkflowID,
			RunID:            dbTask.RunID,
			TaskQueue:        dbTask.TaskQueue,
			ScheduledEventID: dbTask.ScheduledEventID,
		}, nil
	}

	q := e.getOrCreateQueue(queueKey(namespaceID, taskQueue))
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case task := <-q.wfCh:
		return task, nil
	case <-timer.C:
		return nil, nil
	case <-ctx.Done():
		return nil, nil
	}
}

// AddActivityTask adds an activity task to the matching engine.
func (e *Engine) AddActivityTask(ctx context.Context, task *ActivityTask) error {
	q := e.getOrCreateQueue(queueKey(task.NamespaceID, task.TaskQueue))
	select {
	case q.actCh <- task:
		return nil
	default:
		// Channel full, persist to DB
		return e.store.CreateActivityTask(ctx, &persistence.ActivityTask{
			NamespaceID:      task.NamespaceID,
			WorkflowID:       task.WorkflowID,
			RunID:            task.RunID,
			TaskQueue:        task.TaskQueue,
			ActivityID:       task.ActivityID,
			ActivityType:     task.ActivityType,
			ScheduledEventID: task.ScheduledEventID,
			Input:            task.Input,
		})
	}
}

// PollActivityTask polls for an activity task with a timeout.
func (e *Engine) PollActivityTask(ctx context.Context, namespaceID, taskQueue string, timeout time.Duration) (*ActivityTask, error) {
	// First check DB for buffered tasks
	dbTask, err := e.store.ClaimActivityTask(ctx, namespaceID, taskQueue)
	if err != nil {
		return nil, err
	}
	if dbTask != nil {
		return &ActivityTask{
			ID:               dbTask.ID,
			NamespaceID:      dbTask.NamespaceID,
			WorkflowID:       dbTask.WorkflowID,
			RunID:            dbTask.RunID,
			TaskQueue:        dbTask.TaskQueue,
			ActivityID:       dbTask.ActivityID,
			ActivityType:     dbTask.ActivityType,
			ScheduledEventID: dbTask.ScheduledEventID,
			Input:            dbTask.Input,
		}, nil
	}

	q := e.getOrCreateQueue(queueKey(namespaceID, taskQueue))
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case task := <-q.actCh:
		return task, nil
	case <-timer.C:
		return nil, nil
	case <-ctx.Done():
		return nil, nil
	}
}
