package scheduler

import (
	"testing"

	"reflect"

	"net/http"

	"github.com/SimonRichardson/cmdproxy/pkg/peer"
	"github.com/go-kit/kit/log"
)

func TestModeType(t *testing.T) {
	t.Parallel()

	t.Run("parse sequential", func(t *testing.T) {
		mode, err := ParseModeType(string(ModeTypeSequential))
		if err != nil {
			t.Errorf("expected: nil, actual: %v", err)
		}

		if mode != ModeTypeSequential {
			t.Errorf("expected: %v, actual: %v", ModeTypeSequential, mode)
		}
	})

	t.Run("parse parallel", func(t *testing.T) {
		mode, err := ParseModeType(string(ModeTypeParallel))
		if err != nil {
			t.Errorf("expected: nil, actual: %v", err)
		}

		if mode != ModeTypeParallel {
			t.Errorf("expected: %v, actual: %v", ModeTypeParallel, mode)
		}
	})

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseModeType("bad")
		if err == nil {
			t.Errorf("expected: error, actual: %v", err)
		}
	})
}

func TestScheduler(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	t.Run("register", func(t *testing.T) {
		scheduler := NewScheduler(nil, logger)

		task := NewTask(ModeTypeSequential, 0, "hello", true)
		scheduler.Register(task)

		if task.Status() != TaskStatusTypePending {
			t.Errorf("expected: %v, actual: %v", TaskStatusTypePending, task.Status())
		}
	})

	t.Run("get", func(t *testing.T) {
		scheduler := NewScheduler(nil, logger)

		task := NewTask(ModeTypeSequential, 0, "hello", true)
		scheduler.Register(task)

		if tsk, ok := scheduler.Get(task.ID()); !ok || tsk.ID() != task.ID() {
			t.Errorf("no task found")
		}

		if _, ok := scheduler.Get("bad"); ok {
			t.Errorf("invalid task found")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		scheduler := NewScheduler(nil, logger)

		task := NewTask(ModeTypeSequential, 0, "hello", true)
		scheduler.Register(task)
		scheduler.Cancel(task)

		if task.Status() != TaskStatusTypeCancelled {
			t.Errorf("expected: %v, actual: %v", TaskStatusTypeCancelled, task.Status())
		}
	})

	t.Run("peers", func(t *testing.T) {
		peers := []*peer.Peer{
			peer.NewPeer(http.DefaultClient, "http", "0.0.0.0:0", logger),
		}
		scheduler := NewScheduler(peers, logger)
		if !reflect.DeepEqual(scheduler.Peers(), peers) {
			t.Errorf("expected: true, actual: false")
		}
	})
}
