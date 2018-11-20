package oversight

import (
	"fmt"
	"time"
)

// TreeOption are applied to change the behavior of a Tree.
type TreeOption func(*Tree)

// WithSpecification defines a custom setup to tweak restart tolerance and
// strategy for the instance of oversight.
func WithSpecification(maxR int, maxT time.Duration, strategy Strategy) TreeOption {
	return func(t *Tree) {
		WithRestartIntensity(maxR, maxT)(t)
		WithRestartStrategy(strategy)(t)
	}
}

// WithRestartIntensity defines a custom tolerance for failures in the
// supervisor tree.
func WithRestartIntensity(maxR int, maxT time.Duration) TreeOption {
	return func(t *Tree) {
		t.maxR, t.maxT = maxR, maxT
	}
}

// NeverHalt will configure the oversight tree to never stop in face of failure.
func NeverHalt() TreeOption {
	return func(t *Tree) {
		t.maxR = -1
	}
}

// Default restart intensity expectations.
const (
	DefaultMaxR = 1
	DefaultMaxT = 5 * time.Second
)

// DefaultRestartIntensity redefines the tolerance for failures in the
// supervisor tree. It defaults to 1 restart (maxR) in the preceding 5 seconds
// (maxT).
func DefaultRestartIntensity() TreeOption {
	return func(t *Tree) {
		t.maxR, t.maxT = DefaultMaxR, DefaultMaxT
	}
}

// WithRestartStrategy defines a custom restart strategy for the supervisor
// tree.
func WithRestartStrategy(strategy Strategy) TreeOption {
	return func(t *Tree) {
		t.strategy = strategy
	}
}

// DefaultRestartStrategy redefines the supervisor behavior to use OneForOne.
func DefaultRestartStrategy() TreeOption {
	return func(t *Tree) {
		t.strategy = OneForOne()
	}
}

// Processes plugs one or more Permanent child processes to the supervisor tree.
// Processes never reset the child process list.
func Processes(processes ...ChildProcess) TreeOption {
	return func(t *Tree) {
		for _, p := range processes {
			Process(ChildProcessSpecification{
				Restart: Permanent(),
				Start:   p,
			})(t)
		}
	}
}

// Process plugs one child processes to the supervisor tree. Process never reset
// the child process list.
func Process(spec ChildProcessSpecification) TreeOption {
	return func(t *Tree) {
		id := len(t.processes) + 1
		t.states = append(t.states, state{
			stop: func() {
				t.logger.Println("stopped before start")
			},
		})
		if spec.Name == "" {
			spec.Name = fmt.Sprintf("childproc %d", id)
		}
		if _, ok := t.processIndex[spec.Name]; ok {
			spec.Name += fmt.Sprint(" ", id)
		}
		if spec.Restart == nil {
			spec.Restart = Permanent()
		}
		if spec.Shutdown == nil {
			spec.Shutdown = Timeout(DefaultChildProcessTimeout)
		}
		if spec.Start == nil {
			panic("child process must always have a function")
		}
		t.processes = append(t.processes, spec)
		t.processIndex[spec.Name] = id - 1
	}
}

// Logger defines the interface for any logging facility to be compatible with
// oversight trees.
type Logger interface {
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

// WithLogger plugs a custom logger to the oversight tree.
func WithLogger(logger Logger) TreeOption {
	return func(t *Tree) {
		t.logger = logger
	}
}

// WithTree is a shortcut to add a tree as a child process.
func WithTree(subTree *Tree) TreeOption {
	return func(t *Tree) {
		Process(ChildProcessSpecification{
			Restart:  Permanent(),
			Start:    subTree.Start,
			Shutdown: Infinity(),
		})(t)
	}
}