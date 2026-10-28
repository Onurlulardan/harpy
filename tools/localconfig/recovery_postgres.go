package main

func addRecoveryPostgresArtifacts(set *artifactSet) error {
	backup, err := distinctRecoveryPassword(set, nil)
	if err != nil {
		return err
	}
	restore, err := distinctRecoveryPassword(set, map[string]bool{backup: true})
	if err != nil {
		return err
	}
	files := map[string]string{
		"postgres-backup-password.txt": backup,
		"postgres-backup-dsn.txt": postgresDSN(
			"harpy_backup", backup, "harpy"),
		"postgres-restore-password.txt": restore,
		"postgres-restore-dsn.txt": postgresDSN(
			"harpy_restore", restore, "harpy"),
	}
	for name, value := range files {
		if err := set.add(name, []byte(value)); err != nil {
			return err
		}
	}
	return nil
}

func distinctRecoveryPassword(
	set *artifactSet,
	additional map[string]bool,
) (string, error) {
	used := map[string]bool{
		string(set.files["postgres-runtime-password.txt"]):  true,
		string(set.files["postgres-migrator-password.txt"]): true,
	}
	for value := range additional {
		used[value] = true
	}
	for {
		value, err := randomBase64(32)
		if err != nil {
			return "", err
		}
		if !used[value] {
			return value, nil
		}
	}
}
