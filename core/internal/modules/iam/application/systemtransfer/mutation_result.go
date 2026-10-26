package systemtransfer

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func mutationResultIdentity(
	result publicresource.MutationResult,
) (uuid.UUID, int64, error) {
	if result.Validate() != nil {
		return uuid.Nil, 0, publicresource.ErrUnavailable
	}
	read, err := result.Read()
	if err != nil {
		return uuid.Nil, 0, publicresource.ErrUnavailable
	}
	record, err := read.Record()
	if err != nil {
		return uuid.Nil, 0, publicresource.ErrUnavailable
	}
	id, _, err := record.Identity()
	if err != nil {
		return uuid.Nil, 0, publicresource.ErrUnavailable
	}
	receipt, err := result.Receipt()
	if err != nil {
		return uuid.Nil, 0, publicresource.ErrUnavailable
	}
	resultingRevision, _, err := receipt.Result()
	if err != nil {
		return uuid.Nil, 0, publicresource.ErrUnavailable
	}
	return uuid.UUID(id), resultingRevision, nil
}
