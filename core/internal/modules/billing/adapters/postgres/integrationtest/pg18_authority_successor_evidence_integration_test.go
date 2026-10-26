//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertBillingSuccessorEvidenceEquality(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
) {
	t.Helper()
	tests := []struct {
		name         string
		mutate       string
		restore      string
		mutationCode string
	}{
		{
			"requested_by",
			`UPDATE platform.domain_cutovers SET requested_by=
'01900000-0000-7000-8000-000000000093'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			`UPDATE platform.domain_cutovers SET requested_by=
'01900000-0000-7000-8000-000000000091'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			"",
		},
		{
			"approved_by",
			`UPDATE platform.domain_cutovers SET approved_by=
'01900000-0000-7000-8000-000000000094'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			`UPDATE platform.domain_cutovers SET approved_by=
'01900000-0000-7000-8000-000000000092'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			"",
		},
		{
			"reason",
			`UPDATE platform.domain_cutovers SET reason='different reason'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			`UPDATE platform.domain_cutovers SET reason='billing authority cutover'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			"",
		},
		{
			"change_reference",
			`UPDATE platform.domain_cutovers SET change_reference='different:test'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			`UPDATE platform.domain_cutovers SET change_reference='billing-cutover:test'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			"",
		},
		{
			"changed_at",
			`UPDATE platform.domain_cutovers SET changed_at=changed_at+interval '1 second'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			`UPDATE platform.domain_cutovers SET changed_at='2026-07-29T00:00:00Z'
WHERE scope_kind='global' AND scope_id=$1 AND domain='billing'`,
			"",
		},
		{
			"approved_by null",
			`UPDATE platform.cutover_groups SET approved_by=NULL
WHERE scope_kind='global' AND scope_id=$1 AND group_id='billing'`,
			`UPDATE platform.cutover_groups SET approved_by=
'01900000-0000-7000-8000-000000000092'
WHERE scope_kind='global' AND scope_id=$1 AND group_id='billing'`,
			"",
		},
		{
			"approver separation",
			`UPDATE platform.cutover_groups SET approved_by=requested_by
WHERE scope_kind='global' AND scope_id=$1 AND group_id='billing'`,
			`UPDATE platform.cutover_groups SET approved_by=
'01900000-0000-7000-8000-000000000092'
WHERE scope_kind='global' AND scope_id=$1 AND group_id='billing'`,
			"23514",
		},
	}
	for _, test := range tests {
		_, mutationErr := database.ExecContext(
			ctx, test.mutate, billingEnvironmentID,
		)
		if test.mutationCode != "" {
			requireBillingPGCode(
				t, mutationErr, test.mutationCode,
				"cutover rejects "+test.name,
			)
			continue
		}
		if mutationErr != nil {
			t.Fatalf("mutate Billing successor %s: %v", test.name, mutationErr)
		}
		_, callErr := callBillingSuccessor(
			ctx, connection, billingGlobalSuccessorOp,
			billingGlobalSuccessorRun, nil, nil, billingAcceptanceC,
		)
		if _, err := database.ExecContext(
			ctx, test.restore, billingEnvironmentID,
		); err != nil {
			t.Fatalf("restore Billing successor %s: %v", test.name, err)
		}
		requireBillingPGCode(
			t, callErr, "P5721", "successor mismatched "+test.name,
		)
	}
}
