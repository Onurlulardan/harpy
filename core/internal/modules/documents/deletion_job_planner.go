package documents

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var deletionEffectNamespace = uuid.MustParse("533bceca-89a1-4b95-8135-c485e90f19fc")

type deletionPlanner struct{}

func (deletionPlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	parsed, err := jobcontract.ParseDeletion(claim.PayloadReference())
	tenantID, tenantScoped := claim.TenantID()
	_, digest, digestErr := jobcontract.Deletion(parsed.TenantID, parsed.ResourceID)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	wantKey := "documents:blob-delete:" + parsed.ResourceID.String()
	if err != nil || digestErr != nil || !tenantScoped ||
		parsed.TenantID.String() != tenantID.String() ||
		claim.Kind() != asyncmanifestv1.DocumentsBlobDeleteJobKind ||
		claim.PayloadRevision() != 1 || claim.PayloadSHA256() != digest ||
		!hasEffectKey || effectKey != wantKey {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("Documents deletion claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(deletionEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID,
		"documents.blob", "delete", claim.JobID().String(), claim.JobID())
}
