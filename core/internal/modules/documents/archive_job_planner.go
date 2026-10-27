package documents

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var archiveEffectNamespace = uuid.MustParse("82a3ee1a-0589-4457-9e11-7bfd04e9db0b")

type archivePlanner struct{}

func (archivePlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	parsed, err := jobcontract.ParseArchive(claim.PayloadReference())
	tenantID, tenantScoped := claim.TenantID()
	_, digest, digestErr := jobcontract.Archive(parsed.TenantID, parsed.ResourceID)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	wantKey := "documents:archive:" + parsed.ResourceID.String()
	if err != nil || digestErr != nil || !tenantScoped ||
		parsed.TenantID.String() != tenantID.String() ||
		claim.Kind() != asyncmanifestv1.DocumentsArchiveBuildJobKind ||
		claim.PayloadRevision() != 1 || claim.PayloadSHA256() != digest ||
		!hasEffectKey || effectKey != wantKey {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("Documents archive claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(archiveEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID,
		"documents.archive", "build", claim.JobID().String(), claim.JobID())
}
