package platformiampostgres

import (
	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	"github.com/google/uuid"
)

func scanOwnerTransferPrepared(row scanner) (platformdomain.OwnerTransferPrepared, error) {
	var operation, binding, scope uuid.UUID
	var previous, state, proposal []byte
	var result platformdomain.OwnerTransferPrepared
	if err := row.Scan(&operation, &binding, &scope, &result.Generation,
		&previous, &state, &proposal, &result.Status, &result.Replayed); err != nil {
		return result, storeError(err)
	}
	if !copyDigest(&result.PreviousHeadSHA256, previous) ||
		!copyDigest(&result.StateSHA256, state) || !copyDigest(&result.ProposalSHA256, proposal) {
		return platformdomain.OwnerTransferPrepared{}, platformdomain.ErrStorage
	}
	result.OperationID, result.RegistryBindingID = domainID(operation), domainID(binding)
	result.ScopeID = domainID(scope)
	return result, nil
}

func scanOwnerTransferCommitted(row scanner) (platformdomain.OwnerTransferCommitted, error) {
	var operation, predecessor, successor uuid.UUID
	var result platformdomain.OwnerTransferCommitted
	if err := row.Scan(&operation, &predecessor, &successor,
		&result.PredecessorTokenVersion, &result.PredecessorAuthorizationVersion,
		&result.SuccessorTokenVersion, &result.SuccessorAuthorizationVersion,
		&result.OwnerPolicyRevision, &result.PlatformClock, &result.HeadGeneration,
		&result.Replayed); err != nil {
		return result, storeError(err)
	}
	result.OperationID, result.PredecessorUserID = domainID(operation), domainID(predecessor)
	result.SuccessorUserID = domainID(successor)
	return result, nil
}

func scanOwnerTransferCompensated(row scanner) (platformdomain.OwnerTransferCompensated, error) {
	var operation, original uuid.UUID
	var state, ack []byte
	var result platformdomain.OwnerTransferCompensated
	if err := row.Scan(&operation, &original, &result.Generation, &state, &ack, &result.Replayed); err != nil {
		return result, storeError(err)
	}
	if !copyDigest(&result.StateSHA256, state) || !copyDigest(&result.AckSHA256, ack) {
		return platformdomain.OwnerTransferCompensated{}, platformdomain.ErrStorage
	}
	result.OperationID, result.OriginalOperationID = domainID(operation), domainID(original)
	return result, nil
}

func copyDigest(target *[32]byte, source []byte) bool {
	return len(source) == len(target) && copy(target[:], source) == len(target)
}
