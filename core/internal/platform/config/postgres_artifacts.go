package config

type BootstrapArtifacts struct {
	roles     AbsolutePath
	privilege AbsolutePath
	ownership AbsolutePath
	operation AbsolutePath
}

func loadBootstrapArtifacts(values environmentValues) (BootstrapArtifacts, error) {
	paths, err := loadPostgresArtifactPaths(values, []string{
		"HARPY_ROLES_MANIFEST_PATH", "HARPY_PRIVILEGES_MANIFEST_PATH",
		"HARPY_OWNERSHIP_MANIFEST_PATH", "HARPY_BOOTSTRAP_OPERATION_PATH",
	})
	if err != nil {
		return BootstrapArtifacts{}, err
	}
	return BootstrapArtifacts{
		roles: paths[0], privilege: paths[1], ownership: paths[2], operation: paths[3],
	}, nil
}

func (a BootstrapArtifacts) RolesManifestPath() AbsolutePath      { return a.roles }
func (a BootstrapArtifacts) PrivilegesManifestPath() AbsolutePath { return a.privilege }
func (a BootstrapArtifacts) OwnershipManifestPath() AbsolutePath  { return a.ownership }
func (a BootstrapArtifacts) OperationPath() AbsolutePath          { return a.operation }
func (a BootstrapArtifacts) paths() []AbsolutePath {
	return []AbsolutePath{a.roles, a.privilege, a.ownership, a.operation}
}

type ExtensionArtifacts struct {
	extension AbsolutePath
	ownership AbsolutePath
	operation AbsolutePath
}

func loadExtensionArtifacts(values environmentValues) (ExtensionArtifacts, error) {
	paths, err := loadPostgresArtifactPaths(values, []string{
		"HARPY_EXTENSION_MANIFEST_PATH", "HARPY_OWNERSHIP_MANIFEST_PATH",
		"HARPY_EXTENSION_OPERATION_PATH",
	})
	if err != nil {
		return ExtensionArtifacts{}, err
	}
	return ExtensionArtifacts{extension: paths[0], ownership: paths[1], operation: paths[2]}, nil
}

func (a ExtensionArtifacts) ExtensionManifestPath() AbsolutePath { return a.extension }
func (a ExtensionArtifacts) OwnershipManifestPath() AbsolutePath { return a.ownership }
func (a ExtensionArtifacts) OperationPath() AbsolutePath         { return a.operation }
func (a ExtensionArtifacts) paths() []AbsolutePath {
	return []AbsolutePath{a.extension, a.ownership, a.operation}
}

type MigrationArtifacts struct {
	migration AbsolutePath
	extension AbsolutePath
	roles     AbsolutePath
	privilege AbsolutePath
	ownership AbsolutePath
	bootstrap AbsolutePath
	extend    AbsolutePath
	cleanup   AbsolutePath
}

func loadMigrationArtifacts(values environmentValues) (MigrationArtifacts, error) {
	paths, err := loadPostgresArtifactPaths(values, []string{
		"HARPY_MIGRATION_MANIFEST_PATH", "HARPY_EXTENSION_MANIFEST_PATH",
		"HARPY_ROLES_MANIFEST_PATH", "HARPY_PRIVILEGES_MANIFEST_PATH",
		"HARPY_OWNERSHIP_MANIFEST_PATH", "HARPY_BOOTSTRAP_OPERATION_PATH",
		"HARPY_EXTENSION_OPERATION_PATH", "HARPY_CREDENTIAL_CLEANUP_OPERATION_PATH",
	})
	if err != nil {
		return MigrationArtifacts{}, err
	}
	return MigrationArtifacts{
		migration: paths[0], extension: paths[1], roles: paths[2], privilege: paths[3],
		ownership: paths[4], bootstrap: paths[5], extend: paths[6], cleanup: paths[7],
	}, nil
}

func (a MigrationArtifacts) MigrationManifestPath() AbsolutePath  { return a.migration }
func (a MigrationArtifacts) ExtensionManifestPath() AbsolutePath  { return a.extension }
func (a MigrationArtifacts) RolesManifestPath() AbsolutePath      { return a.roles }
func (a MigrationArtifacts) PrivilegesManifestPath() AbsolutePath { return a.privilege }
func (a MigrationArtifacts) OwnershipManifestPath() AbsolutePath  { return a.ownership }
func (a MigrationArtifacts) BootstrapOperationPath() AbsolutePath { return a.bootstrap }
func (a MigrationArtifacts) ExtensionOperationPath() AbsolutePath { return a.extend }
func (a MigrationArtifacts) CleanupOperationPath() AbsolutePath   { return a.cleanup }
func (a MigrationArtifacts) paths() []AbsolutePath {
	return []AbsolutePath{
		a.migration, a.extension, a.roles, a.privilege, a.ownership,
		a.bootstrap, a.extend, a.cleanup,
	}
}

func loadPostgresArtifactPaths(
	values environmentValues,
	keys []string,
) ([]AbsolutePath, error) {
	paths := make([]AbsolutePath, 0, len(keys))
	for _, key := range keys {
		path, err := loadAbsolutePath(values, key)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	if duplicateReleasePath(paths...) {
		return nil, configError("HARPY_*_MANIFEST_PATH", "artifact paths must be distinct")
	}
	return paths, nil
}

func validateArtifactPathIsolation(base, artifacts []AbsolutePath) error {
	paths := append(append([]AbsolutePath(nil), base...), artifacts...)
	if duplicateReleasePath(paths...) {
		return configError("HARPY_*_PATH", "release, receipt, and artifact paths must be distinct")
	}
	return nil
}
