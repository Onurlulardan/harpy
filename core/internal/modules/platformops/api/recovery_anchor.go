package api

import (
	"context"
	"time"
)

type RecoveryAnchorSubject struct {
	EnvironmentID string
	TenantID      string
	ActorID       string
}

type RecoveryAnchor struct {
	AnchorID         string
	Generation       int64
	CredentialID     []byte
	RPID             string
	UserVerification string
	BackupEligible   bool
	BackupState      bool
	State            string
}

type RecoveryAnchorOptionsCommand struct {
	RecoveryAnchorSubject
	OperationID, AnchorID string
	IdempotencyKey        string
	RequestSHA256         string
}

type RecoveryAnchorRegistrationOptions struct {
	OperationID string
	Challenge   string
	RPID        string
	UserID      string
	ExpiresAt   time.Time
	Replayed    bool
}

type RecoveryAnchorRegisterCommand struct {
	RecoveryAnchorSubject
	OperationID    string
	IdempotencyKey string
	RequestSHA256  string
	CredentialJSON []byte
}

type RecoveryAnchorDeleteCommand struct {
	RecoveryAnchorSubject
	OperationID, AnchorID string
	IdempotencyKey        string
	RequestSHA256         string
}

type RecoveryAnchorOperation struct {
	OperationID string
	AnchorID    string
	Kind        string
	State       string
}

type RecoveryAnchors interface {
	ListRecoveryAnchors(context.Context, RecoveryAnchorSubject) ([]RecoveryAnchor, error)
	CreateRecoveryAnchorOptions(
		context.Context, RecoveryAnchorOptionsCommand,
	) (RecoveryAnchorRegistrationOptions, error)
	RegisterRecoveryAnchor(
		context.Context, RecoveryAnchorRegisterCommand,
	) (RecoveryAnchorOperation, error)
	DeleteRecoveryAnchor(
		context.Context, RecoveryAnchorDeleteCommand,
	) (RecoveryAnchorOperation, error)
}
