package iam

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func identityMutationContract(
	claims accesstoken.Claims,
	authorityOperation string,
	input IdentityMutation,
) (uuid.UUID, [32]byte, error) {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 || input.IdempotencyKey == "" {
		return uuid.Nil, [32]byte{}, ErrIdentityHTTPInvalid
	}
	operationID := identityOperationID(claims, authorityOperation, input.IdempotencyKey)
	requestSHA, err := identityRequestSHA(struct {
		ID               uuid.UUID
		ExpectedRevision int64
	}{input.ID, input.ExpectedRevision})
	return operationID, requestSHA, err
}
