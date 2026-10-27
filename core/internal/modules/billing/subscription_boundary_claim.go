package billing

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func subscriptionBoundaryIntent(
	tenantID uuid.UUID,
	claim platformasync.DatabaseJobClaim,
) (subscriptionapp.JobIntent, error) {
	if claim == nil || claim.Kind() != subscriptionBoundaryJobKind ||
		claim.PayloadRevision() != 1 {
		return subscriptionapp.JobIntent{}, subscriptionapp.ErrRequest
	}
	parts := strings.Split(claim.PayloadReference(), ":")
	if len(parts) != 11 || strings.Join(parts[:5], ":") !=
		"urn:harpy:billing:subscription-boundary:v1" {
		return subscriptionapp.JobIntent{}, subscriptionapp.ErrRequest
	}
	jobID, err := uuid.Parse(parts[5])
	if err != nil || jobID != claim.JobID() {
		return subscriptionapp.JobIntent{}, subscriptionapp.ErrRequest
	}
	effectID, err := uuid.Parse(parts[6])
	if err != nil {
		return subscriptionapp.JobIntent{}, subscriptionapp.ErrRequest
	}
	subscriptionID, err := uuid.Parse(parts[7])
	if err != nil {
		return subscriptionapp.JobIntent{}, subscriptionapp.ErrRequest
	}
	runAtMicro, err := strconv.ParseInt(parts[9], 10, 64)
	if err != nil {
		return subscriptionapp.JobIntent{}, subscriptionapp.ErrRequest
	}
	expectedRevision, err := strconv.ParseInt(parts[10], 10, 64)
	if err != nil {
		return subscriptionapp.JobIntent{}, subscriptionapp.ErrRequest
	}
	intent := subscriptionapp.JobIntent{
		ID: jobID, EffectID: effectID, SubscriptionID: subscriptionID,
		Boundary: subscriptionapp.BoundaryKind(parts[8]),
		RunAt:    time.UnixMicro(runAtMicro).UTC(), ExpectedRevision: expectedRevision,
		PayloadRevision: claim.PayloadRevision(),
	}
	artifact, err := subscriptionBoundaryArtifact(tenantID, intent)
	if err != nil || artifact.PayloadReference != claim.PayloadReference() ||
		hex.EncodeToString(artifact.PayloadSHA256[:]) != claim.PayloadSHA256() {
		return subscriptionapp.JobIntent{}, fmt.Errorf(
			"billing boundary claim artifact differs: %w", subscriptionapp.ErrRequest,
		)
	}
	return intent, nil
}
