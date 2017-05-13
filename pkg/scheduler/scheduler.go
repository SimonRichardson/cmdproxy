package scheduler

import (
	"net/http"
	"sync"
	"time"

	"github.com/SimonRichardson/cmdproxy/pkg/peer"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

// ModeType enumerates the types of potential ways to sent to agents.
type ModeType string

const (
	// ModeTypeSequential serves the agents in a sequential mode.
	ModeTypeSequential ModeType = "sequential"

	// ModeTypeParallel serves the agents in a parallel mode.
	ModeTypeParallel ModeType = "parallel"
)

// ParseModeType takes a string and validates it against known ModeTypes
func ParseModeType(s string) (ModeType, error) {
	switch s {
	case string(ModeTypeSequential):
		return ModeTypeSequential, nil
	case string(ModeTypeParallel):
		return ModeTypeParallel, nil
	default:
		return ModeType(""), errors.New("invalid mode type")
	}
}

type Scheduler struct {
	mutex  sync.Mutex
	peers  []*peer.Peer
	logger log.Logger
	tasks  []*Task
	stop   chan chan struct{}
}

// NewScheduler creates a new scheduler, which allows tasks to be mapped over
// the peer agents.
func NewScheduler(peers []*peer.Peer, logger log.Logger) *Scheduler {
	return &Scheduler{
		mutex:  sync.Mutex{},
		peers:  peers,
		logger: logger,
		tasks:  make([]*Task, 0),
		stop:   make(chan chan struct{}),
	}
}

// Register a task for the scheduler to work on.
func (s *Scheduler) Register(task *Task) {
	s.mutex.Lock()
	task.SetStatus(TaskStatusTypePending)
	s.tasks = append(s.tasks, task)
	s.mutex.Unlock()
}

// Cancel a task, even if it's in mid-flight.
func (s *Scheduler) Cancel(task *Task) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	task.SetStatus(TaskStatusTypeCancelled)
	task.Cancel()
}

// Get a task by an ID.
// If no task is found, then it returns false for the boolean.
func (s *Scheduler) Get(id string) (*Task, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, v := range s.tasks {
		if v.ID() == id {
			return v, true
		}
	}
	return nil, false
}

// Run the scheduler, which in turn will execute the following tasks.
// Tasks scheduled are run as FIFO for both sequential and parallel jobs.
func (s *Scheduler) Run() {
	step := time.NewTicker(100 * time.Millisecond)
	defer step.Stop()

	for {
		select {
		case <-step.C:
			s.step()

		case q := <-s.stop:
			close(q)
			return
		}
	}
}

// Stop the scheduler.
func (s *Scheduler) Stop() {
	q := make(chan struct{})
	s.stop <- q
	<-q
}

// Peers return the underlying peers
func (s *Scheduler) Peers() []*peer.Peer {
	return s.peers
}

func (s *Scheduler) step() {
	var task *Task

	s.mutex.Lock()
	for _, v := range s.tasks {
		if v.Status() == TaskStatusTypePending {
			task = v
			break
		}
	}
	s.mutex.Unlock()

	// Nothing to work on, we're done.
	if task == nil {
		return
	}

	// Depending on the task mode, let's pick which strategy to actually use.
	var strat strategy
	switch task.Mode() {
	case ModeTypeSequential:
		strat = s.sequential
	case ModeTypeParallel:
		strat = s.parallel
	default:
		panic(errors.New("invalid mode type"))
	}

	// Run the strategy over the task.
	strat(task)
}

type strategy func(*Task)

func (s *Scheduler) sequential(task *Task) {
	for i := 0; i < len(s.peers); i++ {
		// Something has changed, before scheduled work or if it's happening
		// mid-flight between requests.
		if task.CancelledOrErrored() {
			return
		}

		var (
			peer     = s.peers[(i+task.ClientID())%len(s.peers)]
			req, err = peer.NewRequest(task.Info())
		)
		if err != nil {
			task.SetStatus(TaskStatusTypeErrored)
			level.Error(s.logger).Log("task", task.ID(), "err", err)
			return
		}
		task.SetStatus(TaskStatusTypeRequesting)
		task.cancelFns = append(task.cancelFns, req.Cancel)

		level.Info(s.logger).Log("task", task.ID(), "request", req.URL())

		// Check that we've got a valid result.
		resp, err := req.Do()
		if err != nil {
			level.Warn(s.logger).Log("err", err)
			if task.FailOnError() {
				task.SetStatus(TaskStatusTypeErrored)
				return
			}
		}
		if resp.StatusCode != http.StatusOK {
			level.Warn(s.logger).Log("status", resp.Status)
			if task.FailOnError() {
				task.SetStatus(TaskStatusTypeErrored)
				return
			}
		}
		level.Debug(s.logger).Log("status", resp.Status)
	}

	task.SetStatus(TaskStatusTypeCompleted)
}

func (s *Scheduler) parallel(task *Task) {

}
