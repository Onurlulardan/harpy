package authorization

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

func (request GlobalPolicyRequest) BindingSHA256() ([sha256.Size]byte, error) {
	if request.Validate() != nil {
		return [sha256.Size]byte{}, errPolicyEvaluationContract
	}
	return request.binding, nil
}

func (request GlobalPolicyRequest) EnvironmentID() (uuid.UUID, error) {
	if request.Validate() != nil {
		return uuid.Nil, errPolicyEvaluationContract
	}
	return request.environmentID, nil
}

func (request GlobalPolicyRequest) ActorID() (uuid.UUID, bool, error) {
	if request.Validate() != nil {
		return uuid.Nil, false, errPolicyEvaluationContract
	}
	return request.actorID, !request.public, nil
}

func (request GlobalPolicyRequest) OperationID() (string, error) {
	if request.Validate() != nil {
		return "", errPolicyEvaluationContract
	}
	return request.operationID, nil
}

func (request GlobalPolicyRequest) SnapshotAt() (time.Time, error) {
	if request.Validate() != nil {
		return time.Time{}, errPolicyEvaluationContract
	}
	return request.snapshotAt, nil
}

func (request GlobalPolicyRequest) Public() (bool, error) {
	if request.Validate() != nil {
		return false, errPolicyEvaluationContract
	}
	return request.public, nil
}

func (request GlobalPolicyRequest) QueryIdentitySHA256() ([sha256.Size]byte, error) {
	if request.Validate() != nil {
		return [sha256.Size]byte{}, errPolicyEvaluationContract
	}
	return request.queryIdentity, nil
}

func (request GlobalPolicyRequest) Owner() (string, error) {
	if request.Validate() != nil {
		return "", errPolicyEvaluationContract
	}
	return request.owner, nil
}

func (request GlobalPolicyRequest) Requirements() ([]Requirement, error) {
	if request.Validate() != nil {
		return nil, errPolicyEvaluationContract
	}
	return append([]Requirement{}, request.requirements...), nil
}
