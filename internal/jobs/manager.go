package jobs

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
)

// Kind represents the type of job.
type Kind string

const (
	// KindBuild represents a build job.
	KindBuild Kind = "build"
	// KindPull represents an image pull job.
	KindPull Kind = "pull"
	// KindPush represents an image push job.
	KindPush Kind = "push"
	// KindLog represents a log streaming job.
	KindLog Kind = "log"
)

// State represents the current state of a job.
type State string

const (
	// StateRunning means the job is currently executing.
	StateRunning State = "running"
	// StateDone means the job completed successfully.
	StateDone State = "done"
	// StateFailed means the job completed with an error.
	StateFailed State = "failed"
	// StateCancelled means the job was cancelled by the user.
	StateCancelled State = "cancelled"
)

// Job represents a background streaming job.
type Job struct {
	ID       string
	Kind     Kind
	Target   string
	State    State
	Started  time.Time
	Ended    time.Time
	Stream   cli.Stream
	Lines    []string
	ExitCode int
	Err      error
}

// Elapsed returns the duration of the job. If the job is still running,
// it returns the time since it started. If the job is done, it returns
// the time between start and end.
func (j *Job) Elapsed(clk clock.Clock) time.Duration {
	if j.State == StateRunning {
		return clk.Now().Sub(j.Started)
	}
	return j.Ended.Sub(j.Started)
}

// Manager manages background streaming jobs.
type Manager struct {
	mu    sync.RWMutex
	jobs  map[string]*Job
	clock clock.Clock
	next  int
}

// New creates a new Manager.
func New(clk clock.Clock) *Manager {
	return &Manager{
		jobs:  make(map[string]*Job),
		clock: clk,
		next:  1,
	}
}

// Submit starts a new job that drains the given stream.
func (m *Manager) Submit(kind Kind, target string, stream cli.Stream) *Job {
	m.mu.Lock()
	id := strconv.Itoa(m.next)
	m.next++
	m.mu.Unlock()

	job := &Job{
		ID:      id,
		Kind:    kind,
		Target:  target,
		State:   StateRunning,
		Started: m.clock.Now(),
		Stream:  stream,
		Lines:   []string{},
	}

	m.mu.Lock()
	m.jobs[id] = job
	m.mu.Unlock()

	// Start goroutine to drain stream
	go m.drainStream(job)

	return job
}

// drainStream reads from the job's stream and accumulates lines.
func (m *Manager) drainStream(job *Job) {
	lineCh := make(chan string, 100)
	done := make(chan struct{})

	// Goroutine to accumulate lines
	go func() {
		for line := range lineCh {
			m.mu.Lock()
			job.Lines = append(job.Lines, line)
			m.mu.Unlock()
		}
		close(done)
	}()

	// Read events
	for event := range job.Stream.Events {
		var line string
		switch e := event.(type) {
		case cli.LogLine:
			line = e.Raw
		case cli.RawLine:
			line = e.Text
		case cli.BuildStepEvent:
			line = fmt.Sprintf("#%d [%s] %s", e.Index, e.Stage, e.Step)
		case cli.LayerProgressEvent:
			// For layer events, just track the count
			line = fmt.Sprintf("%d layers", len(e.Layers))
		default:
			line = fmt.Sprintf("%T: %+v", event, event)
		}
		lineCh <- line
	}

	close(lineCh)
	<-done

	// Read result
	result := <-job.Stream.Done

	m.mu.Lock()
	job.Ended = m.clock.Now()
	job.ExitCode = result.ExitCode
	job.Err = result.Err
	if result.Err != nil {
		job.State = StateFailed
	} else {
		job.State = StateDone
	}
	m.mu.Unlock()
}

// Cancel cancels a running job.
func (m *Manager) Cancel(id string) error {
	m.mu.RLock()
	job, ok := m.jobs[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("job %s not found", id)
	}

	if job.State != StateRunning {
		return fmt.Errorf("job %s is not running", id)
	}

	// Cancel the stream
	job.Stream.Cancel()

	m.mu.Lock()
	job.State = StateCancelled
	job.Ended = m.clock.Now()
	m.mu.Unlock()

	return nil
}

// List returns copies of all jobs.
func (m *Manager) List() []*Job {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]*Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobCopy := *job
		jobCopy.Lines = make([]string, len(job.Lines))
		copy(jobCopy.Lines, job.Lines)
		jobs = append(jobs, &jobCopy)
	}
	return jobs
}

// Get returns a copy of a job by ID.
func (m *Manager) Get(id string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[id]
	if !ok {
		return nil, false
	}

	// Return a copy to avoid race conditions
	jobCopy := *job
	jobCopy.Lines = make([]string, len(job.Lines))
	copy(jobCopy.Lines, job.Lines)
	return &jobCopy, true
}
