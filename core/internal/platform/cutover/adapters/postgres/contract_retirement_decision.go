package postgres

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

func (database *ContractRetirementDatabase) retirementDecision(
	ctx context.Context, query string, batch contractretirement.Batch,
) (contractretirement.Decision, error) {
	identity, values, err := batch.DatabaseValues()
	if err != nil {
		return contractretirement.Decision{}, err
	}
	owner, domain, err := identity.DatabaseValues()
	if err != nil {
		return contractretirement.Decision{}, err
	}
	document, err := json.Marshal(values)
	if err != nil {
		return contractretirement.Decision{}, contractretirement.NewFault(
			contractretirement.FaultDependency,
		)
	}
	rows, err := database.queryer.Query(ctx, query, document, owner, domain)
	if err != nil {
		return contractretirement.Decision{}, classifyContractRetirementError(err)
	}
	defer rows.Close()
	retired := make([]contractretirement.Key, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return contractretirement.Decision{}, classifyContractRetirementError(err)
		}
		key, err := contractretirement.ParseKey(value)
		if err != nil {
			return contractretirement.Decision{}, contractretirement.NewFault(
				contractretirement.FaultDependency,
			)
		}
		retired = append(retired, key)
	}
	if err := rows.Err(); err != nil {
		return contractretirement.Decision{}, classifyContractRetirementError(err)
	}
	decision, err := contractretirement.SealDecision(batch, retired)
	if err != nil {
		return contractretirement.Decision{}, contractretirement.NewFault(
			contractretirement.FaultDependency,
		)
	}
	return decision, nil
}
