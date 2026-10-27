package config

import (
	"os"
	"runtime"
	"testing"
)

func TestDatabaseBackupExecutorLoadsIsolatedProfile(t *testing.T) {
	loaded, err := loadDatabaseBackupExecutor(validDatabaseBackupEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.LongRunning().Runtime().Postgres().DSN().Value() == "" ||
		loaded.ManifestSigningKey().Access() != ArtifactReadOnly {
		t.Fatal("backup executor profile contract drifted")
	}
}

func TestDatabaseRestoreLoadsBothExactBranches(t *testing.T) {
	existing, err := loadDatabaseRestore(validDatabaseRestoreEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if existing.Artifacts().Branch() != RestoreExistingReadable ||
		existing.Artifacts().ReceiptPath().Access() != ArtifactAtomicWrite ||
		existing.Runtime().ControlPostgres().DSN().Value() ==
			existing.Runtime().ScratchPostgres().DSN().Value() {
		t.Fatal("existing-readable restore branch drifted")
	}
	rehydrate, err := loadDatabaseRestore(validDatabaseRehydrateEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if rehydrate.Artifacts().Branch() != RestoreRehydrate ||
		rehydrate.Artifacts().RecoveryKit().Access() != ArtifactReadOnceCleanup {
		t.Fatal("rehydrate restore branch drifted")
	}
}

func TestDatabaseRestoreRejectsBranchMixAndRuntimeDSN(t *testing.T) {
	entries := validDatabaseRestoreEnvironment()
	directory := phase14FixtureDirectory("restore-mixed")
	entries = append(entries,
		"HARPY_REHYDRATE_REQUEST_PATH="+phase14FixtureFile(directory, "request.json", false),
		"HARPY_RECOVERY_KIT_FILE="+phase14FixtureFile(directory, "kit", true),
	)
	if _, err := loadDatabaseRestore(entries); err == nil {
		t.Fatal("combined restore branches were accepted")
	}
	if _, err := loadDatabaseRestore(append(validDatabaseRestoreEnvironment(),
		"HARPY_POSTGRES_DSN=postgresql://harpy_restore@example.invalid/db?sslmode=disable")); err == nil {
		t.Fatal("restore accepted normal runtime DSN")
	}
	sameDatabase := validDatabaseRestoreEnvironment()
	scratchPath := environmentEntryValue(
		sameDatabase, "HARPY_RESTORE_SCRATCH_POSTGRES_DSN_FILE",
	)
	if err := os.WriteFile(scratchPath, []byte(
		"postgresql://harpy_restore@example.invalid/harpy?sslmode=disable\n",
	), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDatabaseRestore(sameDatabase); err == nil {
		t.Fatal("restore accepted the live database as scratch")
	}
}

func TestSensitivePhase14FileRequiresOwnerOnlyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission contract")
	}
	entries := validDatabaseBackupEnvironment()
	path := environmentEntryValue(entries, "HARPY_RECOVERY_MANIFEST_SIGNING_KEY_FILE")
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDatabaseBackupExecutor(entries); err == nil {
		t.Fatal("broad private-key permissions were accepted")
	}
}
