package recoveryanchor

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid     = errors.New("recovery anchor request is invalid")
	ErrDenied      = errors.New("recovery anchor operation is denied")
	ErrNotFound    = errors.New("recovery anchor was not found")
	ErrConflict    = errors.New("recovery anchor state conflicts")
	ErrUnavailable = errors.New("recovery anchor dependency is unavailable")
)

type Subject struct {
	EnvironmentID string
	TenantID      string
	ActorID       string
}

type Anchor struct {
	AnchorID         string
	Generation       int64
	CredentialID     []byte
	RPID             string
	UserVerification string
	BackupEligible   bool
	BackupState      bool
	State            string
}

type OptionsCommand struct {
	Subject
	OperationID, AnchorID string
	IdempotencyKey        string
	RequestSHA256         string
}

type RegistrationOptions struct {
	OperationID string
	Challenge   string
	RPID        string
	UserID      string
	ExpiresAt   time.Time
	Replayed    bool
}

type RegisterCommand struct {
	Subject
	OperationID    string
	IdempotencyKey string
	RequestSHA256  string
	CredentialJSON []byte
}

type DeleteCommand struct {
	Subject
	OperationID, AnchorID string
	IdempotencyKey        string
	RequestSHA256         string
}

type Operation struct {
	OperationID string
	AnchorID    string
	Kind        string
	State       string
}

type Service interface {
	List(context.Context, Subject) ([]Anchor, error)
	CreateOptions(context.Context, OptionsCommand) (RegistrationOptions, error)
	Register(context.Context, RegisterCommand) (Operation, error)
	Delete(context.Context, DeleteCommand) (Operation, error)
}
