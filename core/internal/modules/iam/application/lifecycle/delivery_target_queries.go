package lifecycle

import "github.com/google/uuid"

type EmailDeliveryTargetQuery struct {
	Evidence     MutationEvidence
	IdentifierID uuid.UUID
}

type PasswordDeliveryTargetQuery struct {
	IdentifierKind   string
	IdentifierSHA256 [32]byte
}

type PortalDeliveryTargetQuery struct {
	Evidence  MutationEvidence
	BindingID uuid.UUID
}
