package documents

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var adoptionCopyEffectNamespace = uuid.MustParse("7f663d32-8782-43d0-8b64-25a921450c90")

type adoptionCopyPlanner struct{}

func (adoptionCopyPlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	parsed, err := jobcontract.ParseAdoptionCopy(claim.PayloadReference())
	tenantID, tenantScoped := claim.TenantID()
	_, digest, digestErr := jobcontract.AdoptionCopy(parsed.TenantID,
		parsed.ResourceID, parsed.SourceVersionID, parsed.CutoverGeneration,
		parsed.ManifestSHA256)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	wantKey := adoptionCopyKey(parsed.ResourceID.String(), parsed.SourceVersionID.String())
	if err != nil || digestErr != nil || !tenantScoped ||
		parsed.TenantID.String() != tenantID.String() ||
		claim.Kind() != asyncmanifestv1.DocumentsAdoptionCopyJobKind ||
		claim.PayloadRevision() != 1 || claim.PayloadSHA256() != digest ||
		!hasEffectKey || effectKey != wantKey {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("Documents adoption copy claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(adoptionCopyEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID,
		"documents.adoption", "copy", claim.JobID().String(), claim.JobID())
}

func adoptionCopyKey(runID, sourceVersionID string) string {
	return "documents:adoption-copy:" + runID + ":" + sourceVersionID
}
