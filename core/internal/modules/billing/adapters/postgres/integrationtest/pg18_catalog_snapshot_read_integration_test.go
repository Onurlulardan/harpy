//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertCatalogSnapshotReads(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	offering offeringResult,
	discount, campaign catalogResult,
	at time.Time,
) {
	t.Helper()
	cases := []struct {
		adapter string
		file    string
		columns int
		want    uuid.UUID
		args    []any
	}{
		{
			"catalogquery", "list_offerings.sql", 23, offering.offeringID,
			[]any{"", at, 10},
		},
		{
			"catalogquery", "get_offering.sql", 23, offering.offeringID,
			[]any{"rest-public", at},
		},
		{
			"catalogquery", "get_platform_offering.sql", 25, offering.offeringID,
			[]any{offering.offeringID},
		},
		{
			"catalogquery", "get_platform_campaign.sql", 19, campaign.aggregateID,
			[]any{campaign.aggregateID},
		},
		{
			"catalogquery", "get_active_campaign.sql", 18, campaign.aggregateID,
			[]any{at},
		},
	}
	for _, testCase := range cases {
		statement := loadBillingAdapterSQL(t, testCase.adapter, testCase.file)
		got := scanBillingSnapshotID(
			t, ctx, transaction, statement, testCase.columns, testCase.args...,
		)
		if got != testCase.want {
			t.Fatalf("%s id = %s, want %s", testCase.file, got, testCase.want)
		}
	}

	var discountDocument []byte
	if err := transaction.QueryRowContext(
		ctx,
		loadBillingAdapterSQL(t, "catalogquery", "get_platform_discount.sql"),
		discount.aggregateID,
	).Scan(&discountDocument); err != nil || len(discountDocument) <= 2 {
		t.Fatalf("platform discount document length=%d err=%v",
			len(discountDocument), err)
	}

	var legacyID, canonicalID uuid.UUID
	var kind string
	if err := transaction.QueryRowContext(
		ctx,
		loadBillingAdapterSQL(t, "catalogquery", "resolve_legacy_identity.sql"),
		"campaign", campaign.aggregateID,
	).Scan(&legacyID, &canonicalID, &kind); err != nil {
		t.Fatal(err)
	}
	if legacyID != campaign.aggregateID || canonicalID != campaign.aggregateID ||
		kind != "campaign" {
		t.Fatalf("campaign identity = %s/%s/%s", legacyID, canonicalID, kind)
	}
}

func scanBillingSnapshotID(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	statement string,
	columnCount int,
	arguments ...any,
) uuid.UUID {
	t.Helper()
	var identifier uuid.UUID
	values := make([]any, columnCount-1)
	destinations := make([]any, 0, columnCount)
	destinations = append(destinations, &identifier)
	for index := range values {
		destinations = append(destinations, &values[index])
	}
	if err := transaction.QueryRowContext(
		ctx, statement, arguments...,
	).Scan(destinations...); err != nil {
		t.Fatal(err)
	}
	return identifier
}
