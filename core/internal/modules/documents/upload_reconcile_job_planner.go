package documents

import (
	"context"
	"fmt"
	"strconv"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/google/uuid"
)

var uploadReconcileEffectNamespace = uuid.MustParse("e8b71c73-9df3-4cf1-a780-b7afc4c8f268")

type uploadReconcilePlanner struct{}

func (uploadReconcilePlanner) Plan(_ context.Context,
	claim asyncpostgres.JobClaim,
) (asyncpostgres.ExternalEffectRequestSpec, error) {
	parsed, err := jobcontract.ParseUploadReconcile(claim.PayloadReference())
	tenantID, tenantScoped := claim.TenantID()
	_, digest, digestErr := jobcontract.UploadReconcile(parsed.TenantID,
		parsed.ResourceID, parsed.Generation)
	effectKey, hasEffectKey := claim.EffectIdempotencyKey()
	wantKey := uploadReconcileKey(parsed.ResourceID.String(), parsed.Generation)
	if err != nil || digestErr != nil || !tenantScoped ||
		parsed.TenantID.String() != tenantID.String() ||
		claim.Kind() != asyncmanifestv1.DocumentsUploadReconcileJobKind ||
		claim.PayloadRevision() != 1 || claim.PayloadSHA256() != digest ||
		!hasEffectKey || effectKey != wantKey {
		return asyncpostgres.ExternalEffectRequestSpec{},
			fmt.Errorf("Documents upload reconcile claim is invalid")
	}
	attemptID := claim.AttemptID()
	effectID := uuid.NewSHA1(uploadReconcileEffectNamespace, attemptID[:])
	return asyncpostgres.NewExternalEffectRequestSpec(effectID,
		"documents.upload", "reconcile", claim.JobID().String(), claim.JobID())
}

func uploadReconcileKey(uploadID string, generation uint64) string {
	return "documents:upload-reconcile:" + uploadID + ":" +
		strconv.FormatUint(generation, 10)
}
