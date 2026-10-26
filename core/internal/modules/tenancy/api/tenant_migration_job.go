package api

import "time"

type TenantMigrationJob struct {
	JobID                    string
	OperationID              string
	State                    string
	TenantID                 string
	SchemaName               string
	StepPlanSHA256           string
	ReleaseID                string
	RuntimeGeneration        int64
	DatabaseGeneration       int64
	MigrationManifestSHA256  string
	ErrorCode                string
	CompletionReceiptID      string
	CompletionReceiptSHA256  string
	CompletionEvidenceSHA256 string
	CompletedAt              time.Time
	CreatedAt                time.Time
	DatabaseNow              time.Time
}

func (job TenantMigrationJob) IsValid() bool {
	return validTenantMigrationID(job.JobID) &&
		validTenantMigrationID(job.OperationID) &&
		validTenantMigrationID(job.TenantID) && job.SchemaName != "" &&
		validOnboardingDigest(job.StepPlanSHA256) &&
		validTenantMigrationID(job.ReleaseID) &&
		job.RuntimeGeneration > 0 && job.DatabaseGeneration > 0 &&
		validOnboardingDigest(job.MigrationManifestSHA256) &&
		validTenantMigrationState(job.State) && validMigrationCompletion(job)
}

func validMigrationCompletion(job TenantMigrationJob) bool {
	if !microsecondUTC(job.CreatedAt) || !microsecondUTC(job.DatabaseNow) ||
		job.CreatedAt.After(job.DatabaseNow) {
		return false
	}
	if job.State == "succeeded" {
		return validTenantMigrationID(job.CompletionReceiptID) &&
			validOnboardingDigest(job.CompletionReceiptSHA256) &&
			validOnboardingDigest(job.CompletionEvidenceSHA256) &&
			microsecondUTC(job.CompletedAt) && !job.CompletedAt.After(job.DatabaseNow)
	}
	return job.CompletionReceiptID == "" && job.CompletionReceiptSHA256 == "" &&
		job.CompletionEvidenceSHA256 == "" && job.CompletedAt.IsZero()
}

func validTenantMigrationID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return false
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !((character >= '0' && character <= '9') ||
			(character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return value != "00000000-0000-0000-0000-000000000000"
}

func validTenantMigrationState(state string) bool {
	switch state {
	case "pending", "running", "succeeded", "failed",
		"reconcile_required", "cleanup_required", "fenced":
		return true
	default:
		return false
	}
}
