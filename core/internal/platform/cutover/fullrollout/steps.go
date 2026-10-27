package fullrollout

import "errors"

var ErrInvalidEvidence = errors.New("full rollout evidence is invalid")

type ScopeKind string

const (
	ScopeGlobal ScopeKind = "global"
	ScopeTenant ScopeKind = "tenant"
)

type Step struct {
	Scope  ScopeKind
	Domain string
}

var requiredSteps = []Step{
	{Scope: ScopeGlobal, Domain: "billing"},
	{Scope: ScopeTenant, Domain: "iam"},
	{Scope: ScopeTenant, Domain: "billing+forms"},
	{Scope: ScopeTenant, Domain: "workflow"},
	{Scope: ScopeTenant, Domain: "analytics"},
	{Scope: ScopeTenant, Domain: "catalog"},
	{Scope: ScopeTenant, Domain: "work"},
	{Scope: ScopeTenant, Domain: "documents"},
	{Scope: ScopeTenant, Domain: "notifications"},
}

func RequiredSteps() []Step { return append([]Step(nil), requiredSteps...) }

type StepEvidence struct {
	Step                                       Step
	ScopeID, RequestedBy, ApprovedBy           [16]byte
	Generation                                 int64
	State, Writer                              string
	AcceptanceSHA256, ReverseRehearsalSHA256   [32]byte
	SourceCount, TargetCount                   uint64
	SourceSHA256, TargetSHA256                 [32]byte
	PendingJobs, PendingOutbox, PendingEffects uint64
}

func validateSteps(values []StepEvidence) ([16]byte, error) {
	if len(values) != len(requiredSteps) {
		return [16]byte{}, ErrInvalidEvidence
	}
	var tenantID [16]byte
	for index, value := range values {
		if value.Step != requiredSteps[index] || value.ScopeID == [16]byte{} ||
			value.RequestedBy == [16]byte{} || value.ApprovedBy == [16]byte{} ||
			value.RequestedBy == value.ApprovedBy || value.Generation < 1 ||
			value.State != "v2" || value.Writer != "v2" ||
			value.AcceptanceSHA256 == [32]byte{} ||
			value.ReverseRehearsalSHA256 == [32]byte{} ||
			value.SourceCount != value.TargetCount ||
			value.SourceSHA256 == [32]byte{} || value.SourceSHA256 != value.TargetSHA256 ||
			value.PendingJobs != 0 || value.PendingOutbox != 0 || value.PendingEffects != 0 {
			return [16]byte{}, ErrInvalidEvidence
		}
		if value.Step.Scope == ScopeTenant {
			if tenantID == [16]byte{} {
				tenantID = value.ScopeID
			} else if tenantID != value.ScopeID {
				return [16]byte{}, ErrInvalidEvidence
			}
		}
	}
	return tenantID, nil
}
