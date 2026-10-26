package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

const dashboardExportEffectProvider = "analytics.dashboard-export"

var (
	dashboardExportDeleteEffectNamespace = uuid.MustParse("bfce9d42-fd8b-453a-a54e-d6f35a886868")
	dashboardExportExpireEffectNamespace = uuid.MustParse("1bce05e3-fc4d-40a9-82e9-f125b12fa10a")
)

type dashboardExportJobClaim interface {
	AttemptID() uuid.UUID
	JobID() uuid.UUID
	TenantID() (uuid.UUID, bool)
	Kind() string
	PayloadRevision() int
	PayloadReference() string
	PayloadSHA256() string
	EffectClass() jobsv1.EffectClass
	EffectIdempotencyKey() (string, bool)
}

type dashboardExportJobPlanner struct {
	kind, operation string
	namespace       uuid.UUID
}

func (planner dashboardExportJobPlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	return planner.plan(claim)
}

func (planner dashboardExportJobPlanner) plan(
	claim dashboardExportJobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	if claim == nil || !validDashboardExportJobPlanner(planner) {
		return asyncpostgres.ExternalEffectRequestSpec{}, fmt.Errorf(
			"Analytics dashboard export claim is invalid")
	}
	tenantID, tenantScoped := claim.TenantID()
	exportID, parseErr := uuid.Parse(claim.PayloadReference())
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	digest := sha256.Sum256([]byte(claim.PayloadReference()))
	wantKey := planner.kind + ":" + claim.PayloadReference()
	if parseErr != nil || exportID.Version() != uuid.Version(7) || !tenantScoped ||
		tenantID == uuid.Nil || claim.Kind() != planner.kind || claim.PayloadRevision() != 1 ||
		claim.PayloadSHA256() != hex.EncodeToString(digest[:]) ||
		claim.EffectClass() != jobsv1.EffectExternal || !hasEffectKey ||
		effectKey != wantKey || claim.JobID().Version() != uuid.Version(7) ||
		claim.AttemptID() == uuid.Nil || planner.namespace == uuid.Nil {
		return asyncpostgres.ExternalEffectRequestSpec{}, fmt.Errorf(
			"Analytics dashboard export claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(planner.namespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID,
		dashboardExportEffectProvider, planner.operation,
		claim.JobID().String(), claim.JobID())
}

func validDashboardExportJobPlanner(planner dashboardExportJobPlanner) bool {
	return planner.kind == DashboardExportDeleteJobKind && planner.operation == "delete" &&
		planner.namespace == dashboardExportDeleteEffectNamespace ||
		planner.kind == DashboardExportExpireJobKind && planner.operation == "expire" &&
			planner.namespace == dashboardExportExpireEffectNamespace
}
