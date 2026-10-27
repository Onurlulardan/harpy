package systemtransfer

import "context"

// ExportAttemptLoader resolves immutable owner evidence through a worker-only
// tenant scope before actor-bound command scopes are opened.
type ExportAttemptLoader interface {
	LoadExportAttempt(context.Context, ExportJobClaim) (ExportAttempt, error)
}

func matchExportClaim(attempt ExportAttempt, claim ExportJobClaim) error {
	if attempt.TenantID != claim.TenantID || attempt.JobID != claim.JobID ||
		attempt.ExportID != claim.ExportID || attempt.Resource != claim.Resource {
		return ErrExportStateConflict
	}
	return nil
}
