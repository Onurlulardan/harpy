package iam

import (
	"context"
	"errors"
	"time"

	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var ErrProfileBackfillUnmappable = errors.New(
	"IAM profile schema requires an explicit owner mapping",
)

type profileBackfillPlanCreator interface {
	Create(
		context.Context,
		platformtx.TenantScope,
		profilebackfillapp.CreateCommand,
	) (profilebackfillapp.Result, error)
}

type ProfileBackfillPublicationRuntime struct {
	reader *profilebackfillapp.Reader
	plans  profileBackfillPlanCreator
	clock  func() time.Time
}

type ProfileBackfillPublicationResult struct {
	SchemaID           uuid.UUID
	PlanID             uuid.UUID
	Revision           int64
	AuthorizationClock int64
	Replayed           bool
}

func NewProfileBackfillPublicationRuntime(
	catalogs platformasync.WriteCatalogs,
	clock func() time.Time,
) (*ProfileBackfillPublicationRuntime, error) {
	port, err := profilebackfillpostgres.NewSourceReader()
	if err != nil {
		return nil, err
	}
	reader, err := profilebackfillapp.NewReader(port)
	if err != nil {
		return nil, err
	}
	plans, err := NewProfileBackfillPlanRuntime(catalogs)
	if err != nil {
		return nil, err
	}
	return newProfileBackfillPublicationRuntime(reader, plans, clock)
}

func newProfileBackfillPublicationRuntime(
	reader *profilebackfillapp.Reader,
	plans profileBackfillPlanCreator,
	clock func() time.Time,
) (*ProfileBackfillPublicationRuntime, error) {
	if nilProfileHTTPDependency(reader) || nilProfileHTTPDependency(plans) || clock == nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	return &ProfileBackfillPublicationRuntime{reader: reader, plans: plans, clock: clock}, nil
}

func (runtime *ProfileBackfillPublicationRuntime) Replay(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence profileapp.MutationEvidence,
	targetSchemaID uuid.UUID,
) (ProfileBackfillPublicationResult, bool, error) {
	if runtime == nil || runtime.reader == nil || evidence.Validate() != nil ||
		targetSchemaID == uuid.Nil || scope.ActorID() != evidence.ActorID {
		return ProfileBackfillPublicationResult{}, false, profilebackfillapp.ErrInvalid
	}
	receipt, found, err := runtime.reader.Lookup(
		ctx, scope, profileBackfillEvidence(evidence),
	)
	if err != nil || !found {
		return ProfileBackfillPublicationResult{}, false, err
	}
	if receipt.RequestSHA256 != evidence.RequestSHA256 {
		return ProfileBackfillPublicationResult{}, false, profileapp.ErrConflict
	}
	result := ProfileBackfillPublicationResult{
		SchemaID: targetSchemaID, Revision: receipt.Revision,
		AuthorizationClock: receipt.AuthorizationClock, Replayed: true,
	}
	switch receipt.CommandKey {
	case "iam.profile_schema.publish":
		if receipt.ResultID != targetSchemaID {
			return ProfileBackfillPublicationResult{}, false, profileapp.ErrConflict
		}
	case "iam.profile_backfill.plan_create":
		if receipt.TargetSchemaID != targetSchemaID {
			return ProfileBackfillPublicationResult{}, false, profileapp.ErrConflict
		}
		result.PlanID = receipt.ResultID
	default:
		return ProfileBackfillPublicationResult{}, false, profileapp.ErrConflict
	}
	return result, true, nil
}

func (runtime *ProfileBackfillPublicationRuntime) Originate(
	ctx context.Context,
	scope platformtx.TenantScope,
	source profileapp.Schema,
	target profileapp.Schema,
	evidence profileapp.MutationEvidence,
	migrations []ProfileMigrationMappingInput,
) (ProfileBackfillPublicationResult, bool, error) {
	if !profileBackfillPublicationInput(runtime, scope, source, target, evidence) {
		return ProfileBackfillPublicationResult{}, false, profilebackfillapp.ErrInvalid
	}
	required, err := profileBackfillRequired(source, target, migrations)
	if err != nil || !required {
		return ProfileBackfillPublicationResult{}, false, err
	}
	first, err := runtime.readPage(ctx, scope, source, evidence, uuid.Nil)
	if err != nil || len(first) == 0 {
		return ProfileBackfillPublicationResult{}, false, err
	}
	plan, err := runtime.buildPlan(
		ctx, scope, source, target, evidence, migrations, first,
	)
	if err != nil {
		return ProfileBackfillPublicationResult{}, false, err
	}
	result, err := runtime.plans.Create(ctx, scope, profilebackfillapp.CreateCommand{
		Plan: plan, Evidence: profileBackfillEvidence(evidence),
	})
	if err != nil {
		return ProfileBackfillPublicationResult{}, false, err
	}
	return ProfileBackfillPublicationResult{
		SchemaID: target.ID(), PlanID: result.PlanID, Revision: result.Revision,
		AuthorizationClock: result.AuthorizationClock, Replayed: result.Replayed,
	}, true, nil
}
