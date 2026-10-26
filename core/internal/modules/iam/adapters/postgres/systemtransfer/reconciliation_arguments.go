package systemtransfer

import app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"

func reconciliationClaimArguments(schema string,
	claim app.ExportReconciliationClaim,
) []any {
	return []any{schema, claim.TenantID, claim.ExportID, claim.JobID,
		claim.AttemptID, claim.LeaseOwnerID, claim.LeaseGeneration,
		string(claim.Resource)}
}

func reconciliationObservationArguments(schema string,
	claim app.ExportReconciliationClaim,
	observation app.ExportReconciliationObservation,
) []any {
	arguments := reconciliationClaimArguments(schema, claim)
	return append(arguments, observation.EffectID, observation.EffectReceiptID,
		observation.RequestSHA256[:], observation.StartedAt)
}
