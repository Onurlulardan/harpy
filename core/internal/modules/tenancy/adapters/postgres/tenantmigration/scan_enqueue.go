package tenantmigration

import tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"

func scanEnqueue(row Row) (tenantapi.TenantMigrationJob, error) {
	var job tenantapi.TenantMigrationJob
	var errorCode *string
	err := row.Scan(
		&job.JobID, &job.OperationID, &job.State, &job.TenantID,
		&job.SchemaName, &job.StepPlanSHA256, &job.ReleaseID,
		&job.RuntimeGeneration, &job.DatabaseGeneration,
		&job.MigrationManifestSHA256, &errorCode,
	)
	if errorCode != nil {
		job.ErrorCode = *errorCode
	}
	return job, err
}
