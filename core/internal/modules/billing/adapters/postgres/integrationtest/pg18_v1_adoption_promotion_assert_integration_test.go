//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertBillingV1PromotionLifecycle(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	snapshotAt time.Time,
) {
	t.Helper()
	created := snapshotAt.AddDate(-1, 0, 0)
	updated := created.AddDate(0, 0, 1)
	assertAdoptedPromotion(t, ctx, connection, adoptedPromotionExpectation{
		kind: "discount", sourceID: "74000000-0000-4000-8000-000000000074",
		creatorID: "74100000-0000-4000-8000-000000000074",
		name:      "Save 10", description: "Save ten percent",
		createdAt: created, publishedAt: updated,
	})
	assertAdoptedPromotion(t, ctx, connection, adoptedPromotionExpectation{
		kind: "campaign", sourceID: "73000000-0000-4000-8000-000000000073",
		creatorID: "73100000-0000-4000-8000-000000000073",
		name:      "Launch", description: "Launch campaign",
		createdAt: created, publishedAt: updated,
	})
}

type adoptedPromotionExpectation struct {
	kind, sourceID, creatorID string
	name, description         string
	createdAt, publishedAt    time.Time
}

func assertAdoptedPromotion(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	want adoptedPromotionExpectation,
) {
	t.Helper()
	query := adoptedDiscountPromotionSQL
	if want.kind == "campaign" {
		query = adoptedCampaignPromotionSQL
	}
	var lifecycle, name, description string
	var creator uuid.UUID
	var available bool
	var createdAt, publishedAt time.Time
	err := connection.QueryRowContext(ctx, query, want.sourceID).Scan(
		&lifecycle, &available, &name, &description,
		&creator, &createdAt, &publishedAt,
	)
	if err != nil || lifecycle != "active" || available ||
		name != want.name || description != want.description ||
		creator.String() != want.creatorID || !createdAt.Equal(want.createdAt) ||
		!publishedAt.Equal(want.publishedAt) {
		t.Fatalf("adopted %s lifecycle=%q available=%t name=%q description=%q creator=%s created=%s published=%s err=%v",
			want.kind, lifecycle, available, name, description, creator,
			createdAt, publishedAt, err)
	}
}

const adoptedDiscountPromotionSQL = `
SELECT root.lifecycle, version_row.available,
       version_row.presentation_name, version_row.presentation_description,
       root.legacy_created_by, root.created_at, version_row.published_at
FROM platform.billing_v1_adoption_mappings AS mapping
JOIN platform.billing_discounts AS root ON root.id = mapping.target_id
JOIN platform.billing_discount_versions AS version_row
  ON version_row.authority_id = root.authority_id
 AND version_row.discount_id = root.id
WHERE mapping.record_kind = 'discount' AND mapping.source_key = $1
ORDER BY version_row.version DESC LIMIT 1`

const adoptedCampaignPromotionSQL = `
SELECT root.lifecycle, version_row.available,
       version_row.presentation_name, version_row.presentation_description,
       root.legacy_created_by, root.created_at, version_row.published_at
FROM platform.billing_v1_adoption_mappings AS mapping
JOIN platform.billing_campaigns AS root ON root.id = mapping.target_id
JOIN platform.billing_campaign_versions AS version_row
  ON version_row.authority_id = root.authority_id
 AND version_row.campaign_id = root.id
WHERE mapping.record_kind = 'campaign' AND mapping.source_key = $1
ORDER BY version_row.version DESC LIMIT 1`
