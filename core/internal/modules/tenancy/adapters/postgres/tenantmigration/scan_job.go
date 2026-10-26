package tenantmigration

import "time"

import tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"

func scanJob(row Row, scanFound bool) (tenantapi.TenantMigrationJob, bool, error) {
	var job tenantapi.TenantMigrationJob
	var errorCode *string
	var receiptID, receiptSHA256, evidenceSHA256 *string
	var completedAt *time.Time
	var found = true
	destinations := []any{
		&job.JobID, &job.OperationID, &job.State, &job.TenantID,
		&job.SchemaName, &job.StepPlanSHA256, &job.ReleaseID,
		&job.RuntimeGeneration, &job.DatabaseGeneration,
		&job.MigrationManifestSHA256, &errorCode, &receiptID,
		&receiptSHA256, &evidenceSHA256, &completedAt, &job.CreatedAt,
		&job.DatabaseNow,
	}
	if scanFound {
		destinations = append(destinations, &found)
	}
	err := row.Scan(destinations...)
	if errorCode != nil {
		job.ErrorCode = *errorCode
	}
	if receiptID != nil {
		job.CompletionReceiptID = *receiptID
		job.CompletionReceiptSHA256 = *receiptSHA256
		job.CompletionEvidenceSHA256 = *evidenceSHA256
		job.CompletedAt = completedAt.UTC()
	}
	job.DatabaseNow = job.DatabaseNow.UTC()
	job.CreatedAt = job.CreatedAt.UTC()
	if err != nil || !found {
		return tenantapi.TenantMigrationJob{}, found, err
	}
	return job, true, nil
}
