package workflowinitiator

import (
	"context"
	"errors"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var ErrInvalid = errors.New("IAM Workflow initiator verification is invalid")

const StartPermission = "workflow.instance.start"

type Request struct {
	TenantID   [16]byte
	ActorID    [16]byte
	Permission string
	VerifiedAt time.Time
}

func (request Request) Validate() error {
	if request.TenantID == [16]byte{} || request.ActorID == [16]byte{} ||
		request.Permission != StartPermission || request.VerifiedAt.IsZero() ||
		request.VerifiedAt.Location() != time.UTC || request.VerifiedAt.Nanosecond()%1000 != 0 {
		return ErrInvalid
	}
	return nil
}

type Evidence struct {
	Request
	AuthorizationVersion uint64
	AuthorizationClock   uint64
	PermissionID         [16]byte
	SHA256               [32]byte
}

type Verifier interface {
	VerifyWorkflowInitiator(
		context.Context,
		platformtx.TenantScope,
		Request,
	) (Evidence, error)
}
