package lifecycle

import (
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/google/uuid"
)

type Change struct {
	OperationID         uuid.UUID
	RequestSHA256       [32]byte
	CommandKey          string
	AggregateKind       string
	AggregateID         uuid.UUID
	ActorID             uuid.UUID
	Revision            int64
	DeliveryArtifactRef string
}

func (change Change) Validate() error {
	if change.OperationID == uuid.Nil || change.RequestSHA256 == [32]byte{} ||
		change.AggregateID == uuid.Nil || change.ActorID == uuid.Nil ||
		change.Revision < 1 || !commandKeys[change.CommandKey] ||
		!aggregateKinds[change.AggregateKind] {
		return ErrInvalid
	}
	deliveryCommand := change.CommandKey == "iam.email_verification.start" ||
		change.CommandKey == "iam.password_reset.start" ||
		change.CommandKey == "iam.portal_setup.start"
	if deliveryCommand != (change.DeliveryArtifactRef != "") {
		return ErrInvalid
	}
	if deliveryCommand {
		if _, err := lifecycledeliveryv1.ParseReference(change.DeliveryArtifactRef); err != nil {
			return ErrInvalid
		}
	}
	return nil
}

var aggregateKinds = map[string]bool{
	"identifier": true, "one_time_token": true, "portal_binding": true,
	"password_credential": true, "authenticator": true, "session": true,
}

var commandKeys = map[string]bool{
	"iam.identifier.add": true, "iam.identifier.verify": true,
	"iam.identifier.tombstone": true, "iam.identifier.retire": true,
	"iam.email_verification.start": true, "iam.password_reset.start": true,
	"iam.password_reset.complete": true, "iam.portal_setup.start": true,
	"iam.portal_setup.complete": true, "iam.portal_approval.decide": true,
	"iam.password.change": true, "iam.authenticator.suspend": true,
	"iam.authenticator.retire": true, "iam.webauthn.enrollment.start": true,
	"iam.webauthn.enrollment.complete": true, "iam.totp.enrollment.start": true,
	"iam.totp.enrollment.complete": true, "iam.recovery_codes.issue": true,
	"iam.session.revoke_one": true, "iam.session.revoke_all": true,
}
