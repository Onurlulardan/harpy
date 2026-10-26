package config

import "testing"

func TestPhase14ProfilesRejectCrossProfileKeys(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		load    func([]string) error
		foreign string
	}{
		{"initialize backup DSN", validOperatorInitializeEnvironment(DeploymentOnPrem), operatorInitializeError, "HARPY_BACKUP_POSTGRES_DSN_FILE=/foreign"},
		{"identity license", validOperatorIdentityEnvironment(), operatorIdentityError, licenseSnapshotKey + "=/foreign"},
		{"license IAM custody", validOperatorLicenseEnvironment(), operatorLicenseError, "HARPY_IAM_KEY_CUSTODY_MANIFEST_PATH=/foreign"},
		{"tenant password", validTenantMigrationEnvironment(), tenantMigrationError, "HARPY_INITIAL_ADMIN_PASSWORD_FILE=/foreign"},
		{"backup runtime DSN", validDatabaseBackupEnvironment(), databaseBackupError, "HARPY_POSTGRES_DSN=secret"},
		{"restore license", validDatabaseRestoreEnvironment(), databaseRestoreError, licenseSnapshotKey + "=/foreign"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.load(append(test.entries, test.foreign)); err == nil {
				t.Fatal("cross-profile key was accepted")
			}
		})
	}
}

func TestPhase14ProfilesRejectBlankAndDuplicateKeys(t *testing.T) {
	blank := withEnvironment(validOperatorIdentityEnvironment(),
		"HARPY_RESTORE_SECURITY_EVIDENCE_PATH=")
	if _, err := loadOperatorIdentityReconcile(blank); err == nil {
		t.Fatal("blank specialized key was accepted")
	}
	duplicate := append(validOperatorLicenseEnvironment(),
		licenseSnapshotKey+"=/must-not-leak")
	if _, err := loadOperatorLicenseReconcile(duplicate); err == nil {
		t.Fatal("duplicate specialized key was accepted")
	}
}

func operatorInitializeError(entries []string) error {
	_, err := loadOperatorInitialize(entries)
	return err
}
func operatorIdentityError(entries []string) error {
	_, err := loadOperatorIdentityReconcile(entries)
	return err
}
func operatorLicenseError(entries []string) error {
	_, err := loadOperatorLicenseReconcile(entries)
	return err
}
func tenantMigrationError(entries []string) error {
	_, err := loadTenantMigrationExecutor(entries)
	return err
}
func databaseBackupError(entries []string) error {
	_, err := loadDatabaseBackupExecutor(entries)
	return err
}
func databaseRestoreError(entries []string) error {
	_, err := loadDatabaseRestore(entries)
	return err
}
