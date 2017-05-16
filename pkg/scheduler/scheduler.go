package scheduler

import (
	"net/http"
	"sync"
	"sync/atomic"
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

// Scheduler runs Tasks against each peer.
type Scheduler struct {
	mutex  sync.Mutex
	peers  []*peer.Peer
	logger log.Logger
	tasks  []*Task
	stop   chan chan struct{}
	clock  *Clock
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
		clock:  NewClock(),
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
	// Scheduler batches up all tasks.
	step := time.NewTicker(1 * time.Millisecond)
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
	// Get a series of tasks to work on.
	var (
		tasks []*Task
	)

	// Lock to make sure that we don't get any more tasks, whilst we're still
	// trying to workout what to work on.
	s.mutex.Lock()
	for _, v := range s.tasks {
		// Initial state says that it's not actively been worked on, nor has it
		// be scheduled.
		if v.Status() == TaskStatusTypeInitial {
			tasks = append(tasks, v)
			v.SetStatus(TaskStatusTypePending)
		}
	}
	s.mutex.Unlock()

	// Nothing to work on, we're done.
	if len(tasks) < 0 {
		return
	}

	// Go execute the various task requests.
	for _, t := range tasks {
		// Wrap the strategies in a simple runner.
		go run(s.peers, t, s.logger)
	}
}

func run(peers []*peer.Peer, task *Task, logger log.Logger) {
	// Depending on the task mode, let's pick which strategy to actually use.
	var strat strategy
	switch task.Mode() {
	case ModeTypeSequential:
		strat = sequential
	case ModeTypeParallel:
		strat = parallel
	default:
		panic(errors.New("invalid mode type"))
	}

	// Run the strategy over the task.
	strat(peers, task, logger)
}

type strategy func([]*peer.Peer, *Task, log.Logger)

func sequential(peers []*peer.Peer, task *Task, logger log.Logger) {
	for i := 0; i < len(peers); i++ {
		// Something has changed, before scheduled work or if it's happening
		// mid-flight between requests.
		if task.CancelledOrErrored() {
			return
		}

		var (
			peer     = peers[(i+task.ClientID())%len(peers)]
			req, err = peer.NewRequest(task.Info())
		)
		if err != nil {
			task.SetStatus(TaskStatusTypeErrored)
			level.Error(logger).Log("task", task.ID(), "err", err)
			return
		}
		task.SetStatus(TaskStatusTypeRequesting)
		task.addCancelFn(req.Cancel)

		level.Debug(logger).Log("task", task.ID(), "request", req.URL())

		// Check that we've got a valid result.
		resp, err := req.Do()
		if err != nil {
			level.Warn(logger).Log("err", err)
			if task.FailOnError() {
				task.SetStatus(TaskStatusTypeErrored)
				return
			}
		}
		if resp == nil {
			return
		}
		if resp.StatusCode != http.StatusOK {
			level.Warn(logger).Log("status", resp.Status)
			if task.FailOnError() {
				task.SetStatus(TaskStatusTypeErrored)
				return
			}
		}
		level.Debug(logger).Log("status", resp.Status)
	}

	// Make sure we only change to completed if we're still requesting.
	if task.Status() == TaskStatusTypeRequesting {
		task.SetStatus(TaskStatusTypeCompleted)
	}
}

func parallel(peers []*peer.Peer, task *Task, logger log.Logger) {
	// Wait for everything
	var wg sync.WaitGroup
	wg.Add(len(peers))

	// Locate if there are any errors.
	errs := make(chan error)

	for i := 0; i < len(peers); i++ {
		// Something has changed, before scheduled work or if it's happening
		// mid-flight between requests.
		if task.CancelledOrErrored() {
			return
		}

		p := peers[(i+task.ClientID())%len(peers)]
		task.SetStatus(TaskStatusTypeRequesting)

		go func(p *peer.Peer, failOnError bool) {
			defer wg.Done()

			req, err := p.NewRequest(task.Info())
			if err != nil {
				level.Error(logger).Log("task", task.ID(), "err", err)
				return
			}
			task.addCancelFn(req.Cancel)

			level.Debug(logger).Log("task", task.ID(), "request", req.URL())

			// Check that we've got a valid result.
			resp, err := req.Do()
			if err != nil {
				level.Warn(logger).Log("err", err)
				if failOnError {
					errs <- err
					return
				}
			}
			if resp == nil {
				return
			}
			if resp.StatusCode != http.StatusOK {
				level.Warn(logger).Log("status", resp.Status)
				if failOnError {
					errs <- err
					return
				}
			}
			level.Debug(logger).Log("status", resp.Status)
		}(p, task.FailOnError())
	}

	go func() { wg.Wait(); close(errs) }()

	for range errs {
		task.SetStatus(TaskStatusTypeErrored)
		return
	}

	// Make sure we only change to completed if we're still requesting.
	if task.Status() == TaskStatusTypeRequesting {
		task.SetStatus(TaskStatusTypeCompleted)
	}
}

// Clock defines a metric for monitoring how many times something occurred.
type Clock struct {
	times int64
}

// NewClock creates a Clock
func NewClock() *Clock {
	return &Clock{0}
}

// Increment a clock timing
func (c *Clock) Increment() {
	atomic.AddInt64(&c.times, 1)
}

// Time returns how much movement the clock has changed.
func (c *Clock) Time() int64 {
	return atomic.LoadInt64(&c.times)
}
