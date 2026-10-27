package transaction

import (
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func rollbackAfter(transaction *postgres.RuntimeTransaction, cause error) error {
	if err := transaction.Rollback(); err != nil && !errors.Is(err, postgres.ErrTransactionClosed) {
		return errors.Join(cause, ErrRollbackFailed)
	}
	return cause
}
