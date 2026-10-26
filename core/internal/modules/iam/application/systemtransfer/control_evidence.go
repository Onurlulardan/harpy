package systemtransfer

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

type ControlEvidence struct {
	TenantID, ActorID                  uuid.UUID
	TokenVersion, AuthorizationVersion int64
	AuthorizationClock                 int64
	Resource                           publicresource.Resource
}

func (value ControlEvidence) Validate() error {
	if value.TenantID == uuid.Nil || value.ActorID == uuid.Nil ||
		value.TokenVersion < 1 || value.AuthorizationVersion < 1 ||
		value.AuthorizationClock < 1 || !value.Resource.Valid() {
		return ErrControlInvalid
	}
	return nil
}

type ControlRead struct {
	Evidence   ControlEvidence
	TransferID uuid.UUID
}

func (value ControlRead) Validate() error {
	if value.Evidence.Validate() != nil || value.TransferID == uuid.Nil ||
		value.TransferID.Version() != 7 {
		return ErrControlInvalid
	}
	return nil
}
