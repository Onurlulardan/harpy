//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"time"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/google/uuid"
)

type billingV1IntegrationStore struct{ connection *sql.Conn }

func newBillingV1IntegrationStore(connection *sql.Conn) billingv1adoptionapp.ScopeStore {
	return &billingV1IntegrationStore{connection: connection}
}

func (store *billingV1IntegrationStore) Register(
	ctx context.Context,
	operation billingv1adoptionapp.Operation,
	_ billingv1adoptionapp.ScopePlan,
	document []byte,
	evidence billingv1adoptionapp.CutoverEvidence,
) (string, error) {
	var state string
	err := store.connection.QueryRowContext(ctx, `
SELECT platform.billing_register_v1_adoption_v1($1,$2,$3,$4,$5,$6,$7)`,
		operation.OperationID, operation.RunID, operation.AuthorityID,
		document, operation.PlanSHA256, evidence.CanonicalDocument,
		evidence.CanonicalEnvelopeSHA,
	).Scan(&state)
	return state, err
}

func (store *billingV1IntegrationStore) Apply(
	ctx context.Context, operation billingv1adoptionapp.Operation,
) (int64, error) {
	var count int64
	err := store.connection.QueryRowContext(ctx,
		"SELECT platform.billing_apply_v1_adoption_v1($1,$2)",
		operation.OperationID, operation.RunID).Scan(&count)
	return count, err
}

func (store *billingV1IntegrationStore) Verify(
	ctx context.Context, operation billingv1adoptionapp.Operation,
) (billingv1adoptionapp.Reconciliation, error) {
	var result billingv1adoptionapp.Reconciliation
	err := store.connection.QueryRowContext(ctx,
		"SELECT * FROM platform.billing_verify_v1_adoption_v1($1,$2)",
		operation.OperationID, operation.RunID).Scan(
		&result.SourceCount, &result.TargetCount,
		&result.SourceSHA256, &result.TargetSHA256,
		&result.MoneyDifference, &result.PeriodDifference,
		&result.CouponDifference, &result.UsageDifference,
	)
	return result, err
}

func (store *billingV1IntegrationStore) RehearseRollback(
	ctx context.Context,
	operation billingv1adoptionapp.Operation,
	receiptID uuid.UUID,
	observedSourceSHA256 string,
	rehearsedAt time.Time,
) (billingv1adoptionapp.RollbackRehearsal, error) {
	var result billingv1adoptionapp.RollbackRehearsal
	err := store.connection.QueryRowContext(ctx, `
SELECT * FROM platform.billing_rehearse_v1_adoption_rollback_v1($1,$2,$3,$4,$5)`,
		operation.OperationID, operation.RunID, receiptID,
		observedSourceSHA256, rehearsedAt,
	).Scan(&result.ReceiptID, &result.SourceSHA256,
		&result.ReverseSHA256, &result.RehearsedAt)
	result.RehearsedAt = result.RehearsedAt.UTC()
	return result, err
}
