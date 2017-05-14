package scheduler

import (
	"testing"
	"testing/quick"

	"github.com/SimonRichardson/cmdproxy/pkg/test"
)

func TestTask(t *testing.T) {
	t.Parallel()

	t.Run("id", func(t *testing.T) {
		fn := func(a int) bool {
			task := NewTask(ModeTypeSequential, a, "info", false)
			return task.ClientID() == a
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("mode", func(t *testing.T) {
		task := NewTask(ModeTypeSequential, 1, "info", false)
		if task.Mode() != ModeTypeSequential {
			t.Errorf("expected: %s, actual: %s", string(ModeTypeSequential), string(task.Mode()))
		}
	})

	t.Run("info", func(t *testing.T) {
		fn := func(a test.ASCII) bool {
			task := NewTask(ModeTypeSequential, 0, a.String(), false)
			return task.Info() == a.String()
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("failOnError", func(t *testing.T) {
		fn := func(a bool) bool {
			task := NewTask(ModeTypeSequential, 0, "info", a)
			return task.FailOnError() == a
		}

		if err := quick.Check(fn, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestTaskStatus(t *testing.T) {
	t.Parallel()

	t.Run("status", func(t *testing.T) {
		task := NewTask(ModeTypeSequential, 1, "info", false)
		task.SetStatus(TaskStatusTypePending)
		if task.Status() != TaskStatusTypePending {
			t.Errorf("expected: %s, actual: %s", string(TaskStatusTypePending), string(task.Status()))
		}

		task.SetStatus(TaskStatusTypeCompleted)
		if task.Status() != TaskStatusTypeCompleted {
			t.Errorf("expected: %s, actual: %s", string(TaskStatusTypeCompleted), string(task.Status()))
		}

		task.SetStatus(TaskStatusTypeErrored)
		if task.Status() != TaskStatusTypeErrored {
			t.Errorf("expected: %s, actual: %s", string(TaskStatusTypeErrored), string(task.Status()))
		}
	})

	t.Run("cancelledOrErrored", func(t *testing.T) {
		task := NewTask(ModeTypeSequential, 1, "info", false)
		task.SetStatus(TaskStatusTypePending)
		if task.CancelledOrErrored() {
			t.Errorf("expected: false, actual: %v", task.CancelledOrErrored())
		}

		task.SetStatus(TaskStatusTypeCancelled)
		if !task.CancelledOrErrored() {
			t.Errorf("expected: true, actual: %v", task.CancelledOrErrored())
		}

		task.SetStatus(TaskStatusTypeErrored)
		if !task.CancelledOrErrored() {
			t.Errorf("expected: true, actual: %v", task.CancelledOrErrored())
		}

		task.SetStatus(TaskStatusTypeRequesting)
		if task.CancelledOrErrored() {
			t.Errorf("expected: false, actual: %v", task.CancelledOrErrored())
		}
	})
}

func TestTaskCancel(t *testing.T) {
	t.Parallel()

	t.Run("cancel", func(t *testing.T) {
		called := false
		task := NewTask(ModeTypeSequential, 1, "info", false)

		task.Cancel()

		if called {
			t.Errorf("expected: false, actual: %v", called)
		}

		task.addCancelFn(func() {
			called = true
		})

		if called {
			t.Errorf("expected: false, actual: %v", called)
		}

		task.Cancel()

		if !called {
			t.Errorf("expected: true, actual: %v", called)
		}
	})
}
