package resolved

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (binding SnapshotBinding) Scope() (SnapshotScope, error) {
	if binding.Validate() != nil {
		return 0, snapshotBindingError()
	}
	return binding.scope, nil
}

func (binding SnapshotBinding) ScopeID() (uuid.UUID, error) {
	if binding.Validate() != nil {
		return uuid.Nil, snapshotBindingError()
	}
	return binding.scopeID, nil
}

func (binding SnapshotBinding) SubjectID() (uuid.UUID, bool, error) {
	if binding.Validate() != nil {
		return uuid.Nil, false, snapshotBindingError()
	}
	return binding.subjectID, binding.subjectID != uuid.Nil, nil
}

func (binding SnapshotBinding) Audience() (SnapshotAudience, error) {
	if binding.Validate() != nil {
		return 0, snapshotBindingError()
	}
	return binding.audience, nil
}

func (binding SnapshotBinding) OperationID() (string, error) {
	if binding.Validate() != nil {
		return "", snapshotBindingError()
	}
	return binding.operationID, nil
}

func (binding SnapshotBinding) SnapshotAt() (time.Time, bool, error) {
	if binding.Validate() != nil {
		return time.Time{}, false, snapshotBindingError()
	}
	return binding.snapshotAt, !binding.snapshotAt.IsZero(), nil
}

func snapshotBindingError() error {
	return fmt.Errorf("REST Query snapshot binding is invalid")
}
