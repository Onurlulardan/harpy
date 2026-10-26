package billing

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	subscriptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/subscription"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type subscriptionBoundaryPayload struct {
	TenantID         string `json:"tenant_id"`
	JobID            string `json:"job_id"`
	EffectID         string `json:"effect_id"`
	SubscriptionID   string `json:"subscription_id"`
	Boundary         string `json:"boundary"`
	RunAt            string `json:"run_at"`
	ExpectedRevision int64  `json:"expected_revision"`
	PayloadRevision  int    `json:"payload_revision"`
}

func subscriptionBoundaryArtifact(
	tenantID uuid.UUID,
	intent subscriptionapp.JobIntent,
) (subscriptionpostgres.BoundaryArtifact, error) {
	document, err := json.Marshal(subscriptionBoundaryPayload{
		TenantID: tenantID.String(), JobID: intent.ID.String(),
		EffectID: intent.EffectID.String(), SubscriptionID: intent.SubscriptionID.String(),
		Boundary: string(intent.Boundary), RunAt: intent.RunAt.Format(time.RFC3339Nano),
		ExpectedRevision: intent.ExpectedRevision,
		PayloadRevision:  intent.PayloadRevision,
	})
	if err != nil {
		return subscriptionpostgres.BoundaryArtifact{}, err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 4096)
	if err != nil {
		return subscriptionpostgres.BoundaryArtifact{}, err
	}
	return subscriptionpostgres.BoundaryArtifact{
		PayloadReference: fmt.Sprintf(
			"urn:harpy:billing:subscription-boundary:v1:%s:%s:%s:%s:%d:%d",
			intent.ID, intent.EffectID, intent.SubscriptionID, intent.Boundary,
			intent.RunAt.UnixMicro(), intent.ExpectedRevision,
		),
		PayloadSHA256: sha256.Sum256(canonical),
	}, nil
}
