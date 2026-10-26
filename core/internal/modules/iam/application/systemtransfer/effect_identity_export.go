package systemtransfer

import (
	"errors"

	"github.com/google/uuid"
)

func (worker *ExportWorker) newExportEffectReceiptID(
	effectID uuid.UUID,
) (uuid.UUID, error) {
	now, err := worker.exportNow()
	if err != nil {
		return uuid.Nil, err
	}
	identity, err := worker.identities.NewExportIdentity(now)
	if err != nil || !exportV7(identity) || identity == effectID {
		return uuid.Nil, errors.Join(ErrExportUnavailable, err)
	}
	return identity, nil
}
