package storagecatalog

import "fmt"

func validateOwner(owner Owner, expectedName string) error {
	if owner.Name != expectedName {
		return fmt.Errorf(
			"storage owner name %q differs from fragment name %q",
			owner.Name, expectedName,
		)
	}
	if owner.Scopes == nil || owner.MigrationDirs == nil ||
		owner.StaticSQLDirs == nil || owner.PlatformSchemas == nil ||
		owner.SchemaUsageGrants == nil || owner.PlatformRoutines == nil ||
		owner.RoutineExecGrants == nil || owner.PlatformTables == nil ||
		owner.ExternalAccesses == nil || owner.ForeignKeys == nil ||
		owner.LegacyTables == nil || owner.DynamicNamespaces == nil {
		return fmt.Errorf("storage owner %q arrays must not be null", owner.Name)
	}
	for _, foreignKey := range owner.ForeignKeys {
		if foreignKey.SourceColumns == nil || foreignKey.TargetColumns == nil {
			return fmt.Errorf(
				"storage owner %q foreign key arrays must not be null", owner.Name,
			)
		}
	}
	return nil
}
