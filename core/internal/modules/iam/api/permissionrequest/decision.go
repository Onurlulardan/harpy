package permissionrequest

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Decision string

const (
	Approve Decision = "APPROVE"
	Reject  Decision = "REJECT"
)

type StepUpEvidence struct {
	PurposeSHA256 [32]byte
	AuthTime      time.Time
	ACR           string
	AMRPassword   bool
	AMRWebAuthn   bool
	AMRTOTP       bool
}

type Command struct {
	OperationID                          string
	RequestSHA256                        [32]byte
	GroupID                              string
	ExpectedRevision                     int64
	ReviewerID                           string
	ExpectedReviewerTokenVersion         int64
	ExpectedReviewerAuthorizationVersion int64
	ReviewTaskID                         string
	Decision                             Decision
	StepUp                               StepUpEvidence
}

type Result struct {
	Code               string
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

type Decider interface {
	DecidePermissionRequest(
		context.Context,
		platformtx.TenantScope,
		Command,
	) (Result, error)
}
