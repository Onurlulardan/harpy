package postgres

import (
	"context"
	"errors"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/jackc/pgx/v5"
)

var (
	ErrDeadLetterStoreUnavailable = errors.New("dead-letter store is unavailable")
	ErrDeadLetterInputInvalid     = errors.New("dead-letter input is invalid")
	ErrDeadLetterAuthority        = errors.New("dead-letter authority is unavailable")
	ErrDeadLetterNotFound         = errors.New("dead-letter root is unavailable")
	ErrDeadLetterPayloadExpired   = errors.New("dead-letter payload retention has ended")
	ErrDeadLetterRetryInvalid     = errors.New("dead-letter source cannot be retried")
)

type deadLetterDatabase interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

type deadLetterRepository struct {
	database  deadLetterDatabase
	cleanup   time.Duration
	consumers eventsv1.ConsumerCatalog
}

func (repository *deadLetterRepository) valid() bool {
	return repository != nil && repository.database != nil && repository.cleanup > 0
}

func (repository *deadLetterRepository) validIngest() bool {
	return repository.valid() && len(repository.consumers.Policies()) > 0
}

var _ deadletter.IngestStore = (*deadLetterRepository)(nil)
var _ deadletter.Store = (*deadLetterRepository)(nil)
