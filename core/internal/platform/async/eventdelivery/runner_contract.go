package eventdelivery

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

var (
	ErrRunnerDependencies = errors.New("event delivery runner dependencies are incomplete")
	ErrRunnerStarted      = errors.New("event delivery runner is already started")
	ErrSessionUnavailable = errors.New("event delivery session is unavailable")
	ErrSessionPanicked    = errors.New("event delivery session panicked")
)

type brokerSession interface {
	Run(context.Context) error
	Close() error
}

type sessionConnector interface {
	Open(context.Context, Registration) (brokerSession, error)
}

type jitterSource interface {
	Uint64N(uint64) uint64
}

type waitFunction func(context.Context, time.Duration) error

type Runner struct {
	registry  Registry
	connector sessionConnector
	admission *admission.Gate
	jitter    jitterSource
	wait      waitFunction
	lifecycle runnerLifecycle
}
