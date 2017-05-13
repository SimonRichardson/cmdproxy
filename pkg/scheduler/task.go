package scheduler

import (
	"context"

	"github.com/pborman/uuid"
)

type Task struct {
	id          string
	mode        ModeType
	clientID    int
	info        string
	failOnError bool
	status      TaskStatusType
	cancelFns   []context.CancelFunc
}

func NewTask(mode ModeType, clientID int, info string, failOnError bool) *Task {
	return &Task{
		id:          uuid.NewUUID().String(),
		mode:        mode,
		clientID:    clientID,
		info:        info,
		failOnError: failOnError,
		status:      TaskStatusTypePending,
	}
}

func (t *Task) ID() string {
	return t.id
}

func (t *Task) Mode() ModeType {
	return t.mode
}

func (t *Task) ClientID() int {
	return t.clientID
}

func (t *Task) Info() string {
	return t.info
}

func (t *Task) FailOnError() bool {
	return t.failOnError
}

func (t *Task) Status() TaskStatusType {
	return t.status
}

func (t *Task) SetStatus(s TaskStatusType) {
	t.status = s
}

// Cancel attempts to cancel any requesting peer agents updates.
// Cancel is idempotent and can be called multiple times.
func (t *Task) Cancel() {
	for _, v := range t.cancelFns {
		v()
	}
}

// CancelledOrErrored determins if the task can continue or if it failed.
func (t *Task) CancelledOrErrored() bool {
	return t.status == TaskStatusTypeCancelled || t.status == TaskStatusTypeErrored
}

// TaskStatusType defines the state of the Task as it proceeds through the
// scheduler.
type TaskStatusType string

const (
	TaskStatusTypePending TaskStatusType = "pending"

	TaskStatusTypeRequesting TaskStatusType = "requesting"

	TaskStatusTypeCompleted TaskStatusType = "completed"

	TaskStatusTypeCancelled TaskStatusType = "cancelled"

	TaskStatusTypeErrored TaskStatusType = "errored"
)
