package invitation

import (
	"strings"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/google/uuid"
)

type AuditChange struct {
	OperationID         uuid.UUID
	ActorID             uuid.UUID
	InvitationID        uuid.UUID
	RequestSHA256       [32]byte
	CommandKey          string
	ResultState         State
	ResultRevision      int64
	DeliveryArtifactRef string
	AuthorizationClock  int64
}

func (change AuditChange) Validate() error {
	if change.OperationID == uuid.Nil || change.ActorID == uuid.Nil ||
		change.InvitationID == uuid.Nil || change.RequestSHA256 == [32]byte{} ||
		change.ResultRevision < 1 || !change.ResultState.Valid() {
		return ErrInvalid
	}
	switch change.CommandKey {
	case "iam.invitation.create":
		if change.ResultState != StatePending || change.AuthorizationClock != 0 {
			return ErrInvalid
		}
		if _, err := deliveryartifactv1.ParseReference(
			change.DeliveryArtifactRef,
		); err != nil {
			return ErrInvalid
		}
	case "iam.invitation.revoke":
		if change.ResultState != StateRevoked || change.DeliveryArtifactRef != "" ||
			change.AuthorizationClock != 0 {
			return ErrInvalid
		}
	case "iam.invitation.expire":
		if change.ResultState != StateExpired || change.DeliveryArtifactRef != "" ||
			change.AuthorizationClock != 0 {
			return ErrInvalid
		}
	case "iam.invitation.accept":
		if change.ResultState != StateAccepted || change.DeliveryArtifactRef != "" ||
			change.AuthorizationClock < 1 {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	if strings.Contains(change.DeliveryArtifactRef, "?") ||
		strings.Contains(change.DeliveryArtifactRef, "#") {
		return ErrInvalid
	}
	return nil
}
