package v1adoption

import (
	"context"
	"fmt"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
)

func (storage *store) Register(
	ctx context.Context,
	operation billingv1adoptionapp.Operation,
	_ billingv1adoptionapp.ScopePlan,
	document []byte,
	evidence billingv1adoptionapp.CutoverEvidence,
) (string, error) {
	var state string
	err := storage.database.QueryRow(
		ctx, registerScopeSQL, operation.OperationID, operation.RunID,
		operation.AuthorityID, document, operation.PlanSHA256,
		evidence.CanonicalDocument, evidence.CanonicalEnvelopeSHA,
	).Scan(&state)
	if err != nil {
		return "", fmt.Errorf("persist Billing V1 adoption scope: %w", err)
	}
	return state, nil
}
