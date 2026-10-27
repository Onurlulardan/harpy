package main

import (
	"fmt"

	corecutover "github.com/Onurlulardan/harpy/core/public/cutover"
)

func decodeFullRolloutSteps(values []fullRolloutStep) (
	[]corecutover.FullRolloutStepEvidence, error,
) {
	result := make([]corecutover.FullRolloutStepEvidence, len(values))
	for index, value := range values {
		scopeID, err := parseFullRolloutUUID(value.ScopeID)
		if err != nil {
			return nil, err
		}
		requestedBy, err := parseFullRolloutUUID(value.RequestedBy)
		if err != nil {
			return nil, err
		}
		approvedBy, err := parseFullRolloutUUID(value.ApprovedBy)
		if err != nil {
			return nil, err
		}
		digests, err := decodeFullRolloutDigests([]string{value.AcceptanceSHA256,
			value.ReverseRehearsalSHA256, value.SourceSHA256, value.TargetSHA256})
		if err != nil {
			return nil, err
		}
		scope := corecutover.FullRolloutScopeKind(value.Scope)
		if scope != corecutover.FullRolloutScopeGlobal &&
			scope != corecutover.FullRolloutScopeTenant {
			return nil, fmt.Errorf("completion evidence scope is invalid")
		}
		result[index] = corecutover.FullRolloutStepEvidence{
			Step:    corecutover.FullRolloutStep{Scope: scope, Domain: value.Domain},
			ScopeID: [16]byte(scopeID), RequestedBy: [16]byte(requestedBy),
			ApprovedBy: [16]byte(approvedBy), Generation: value.Generation,
			State: value.State, Writer: value.Writer, AcceptanceSHA256: digests[0],
			ReverseRehearsalSHA256: digests[1], SourceCount: value.SourceCount,
			TargetCount: value.TargetCount, SourceSHA256: digests[2],
			TargetSHA256: digests[3], PendingJobs: value.PendingJobs,
			PendingOutbox: value.PendingOutbox, PendingEffects: value.PendingEffects,
		}
	}
	return result, nil
}
