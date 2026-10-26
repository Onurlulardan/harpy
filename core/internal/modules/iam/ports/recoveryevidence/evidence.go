package recoveryevidence

import (
	"context"
	"encoding/json"
)

type Anchor struct {
	EnvironmentID           string
	TenantID                string
	OwnerUserID             string
	OwnerBindingID          string
	OwnerBindingGeneration  int64
	AnchorID                string
	AnchorGeneration        int64
	CredentialID            []byte
	PublicKeyCOSE           []byte
	AAGUID                  string
	AttestationFormat       string
	AttestationType         string
	RPID                    string
	AllowedOrigins          json.RawMessage
	AllowedOriginsSHA256    []byte
	UserVerification        string
	BackupEligible          bool
	BackupState             bool
	SignCount               int64
	CounterPolicy           string
	ProtectedHeadGeneration int64
	ProtectedStateSHA256    []byte
	ProtectedAckSHA256      []byte
}

type Reader interface {
	ReadOwnerAnchors(context.Context, string) ([]Anchor, error)
}
