package deadletterdelivery

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDependencies       = errors.New("dead-letter delivery dependencies are incomplete")
	ErrAlreadyStarted     = errors.New("dead-letter delivery workload is already started")
	ErrSessionUnavailable = errors.New("dead-letter delivery session is unavailable")
	ErrSessionPanicked    = errors.New("dead-letter delivery session panicked")
)

type Session interface {
	Run(context.Context) error
	Close() error
}

type Connector interface {
	Open(context.Context) (Session, error)
}

type Settings struct {
	ReconnectMinimum time.Duration
	ReconnectMaximum time.Duration
}

func (settings Settings) valid() bool {
	return settings.ReconnectMinimum > 0 &&
		settings.ReconnectMaximum >= settings.ReconnectMinimum
}
