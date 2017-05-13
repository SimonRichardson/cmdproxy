package scheduler

import (
	"context"

	"sync"

	"github.com/pborman/uuid"
)

// Task collects all the information required for the scheduler, which can then
// be used to find the status of the task with in the lifetime.
type Task struct {
	mutex       sync.Mutex
	id          string
	mode        ModeType
	clientID    int
	info        string
	failOnError bool
	status      TaskStatusType
	cancelFns   []context.CancelFunc
}

// NewTask creates a Task with all the model data.
func NewTask(mode ModeType, clientID int, info string, failOnError bool) *Task {
	return &Task{
		mutex:       sync.Mutex{},
		id:          uuid.NewUUID().String(),
		mode:        mode,
		clientID:    clientID,
		info:        info,
		failOnError: failOnError,
		status:      TaskStatusTypePending,
	}
}

// ID returns the associated ID with the Task.
func (t *Task) ID() string {
	return t.id
}

// Mode defines how the Task should be run, when executed (sequential, parallel)
func (t *Task) Mode() ModeType {
	return t.mode
}

// ClientID defines what offset should be used in the peers when starting the
// execution.
// This offset only matters in sequential mode. When in parallel mode, the
// runtime goroutine scheduler can shuffle the order of execution.
func (t *Task) ClientID() int {
	return t.clientID
}

// Info defines what should be sent to the proxy agents.
func (t *Task) Info() string {
	return t.info
}

// FailOnError prevents execution of subsequent requests when a failure occurs.
func (t *Task) FailOnError() bool {
	return t.failOnError
}

// Status defines the TaskStatusType of the Task.
func (t *Task) Status() TaskStatusType {
	return t.status
}

// SetStatus allows the updating of the status.
func (t *Task) SetStatus(s TaskStatusType) {
	t.status = s
}

// Cancel attempts to cancel any requesting peer agents updates.
// Cancel is idempotent and can be called multiple times.
func (t *Task) Cancel() {
	t.mutex.Lock()
	for _, v := range t.cancelFns {
		v()
	}
	t.mutex.Unlock()
}

// CancelledOrErrored determins if the task can continue or if it failed.
func (t *Task) CancelledOrErrored() bool {
	return t.status == TaskStatusTypeCancelled || t.status == TaskStatusTypeErrored
}

func (t *Task) addCancelFn(fn context.CancelFunc) {
	t.mutex.Lock()
	t.cancelFns = append(t.cancelFns, fn)
	t.mutex.Unlock()
}

// TaskStatusType defines the state of the Task as it proceeds through the
// scheduler.
// Typically you would expect: pending -> requesting -> completed.
type TaskStatusType string

const (
	// TaskStatusTypePending defines the initial state of the task.
	TaskStatusTypePending TaskStatusType = "pending"

	// TaskStatusTypeRequesting labels the Task when it's requesting.
	TaskStatusTypeRequesting TaskStatusType = "requesting"

	// TaskStatusTypeCompleted labels the Task once it's completed.
	TaskStatusTypeCompleted TaskStatusType = "completed"

	// TaskStatusTypeCancelled labels the Task when it's cancelled.
	TaskStatusTypeCancelled TaskStatusType = "cancelled"

	// TaskStatusTypeErrored labels the Task once it's errored.
	TaskStatusTypeErrored TaskStatusType = "errored"
)
