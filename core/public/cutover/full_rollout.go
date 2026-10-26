package cutover

import (
	"context"
	"time"

	cutoverpostgres "github.com/Onurlulardan/harpy/core/internal/platform/cutover/adapters/postgres"
	cutoverapplication "github.com/Onurlulardan/harpy/core/internal/platform/cutover/application"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/fullrollout"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
)

type FullRolloutCompletion = fullrollout.PersistedCompletion
type FullRolloutEvidence = fullrollout.CompletionEvidence
type FullRolloutStep = fullrollout.Step
type FullRolloutStepEvidence = fullrollout.StepEvidence
type FullRolloutScopeKind = fullrollout.ScopeKind
type FullRolloutCohort = grouprehearsal.Cohort
type FullRolloutSoakEvidence = grouprehearsal.SoakEvidence
type FullRolloutMetrics = grouprehearsal.Metrics
type FullRolloutDatabase = cutoverpostgres.FullRolloutDatabase

const (
	FullRolloutScopeGlobal = fullrollout.ScopeGlobal
	FullRolloutScopeTenant = fullrollout.ScopeTenant
)

func EvaluateFullRollout(value FullRolloutEvidence) error {
	return fullrollout.EvaluateCompletion(value)
}

func RequiredFullRolloutSteps() []FullRolloutStep {
	return fullrollout.RequiredSteps()
}

func FullRolloutCohortOrder() []FullRolloutCohort {
	return grouprehearsal.CohortOrder()
}

func FullRolloutMinimumSoak(value FullRolloutCohort) (
	time.Duration, error,
) {
	return grouprehearsal.MinimumSoak(value)
}

func RecordFullRollout(
	ctx context.Context, database FullRolloutDatabase, value FullRolloutCompletion,
) ([32]byte, error) {
	recorder, err := cutoverpostgres.NewFullRolloutRecorder(database)
	if err != nil {
		return [32]byte{}, err
	}
	receipt, err := cutoverapplication.RecordFullRollout(ctx, recorder, value)
	if err != nil {
		return [32]byte{}, err
	}
	return receipt.SHA256()
}
