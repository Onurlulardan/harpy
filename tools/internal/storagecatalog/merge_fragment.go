package storagecatalog

import "fmt"

func mergeOwnerFragment(owner *Owner, fragment ownerFragmentDocument) error {
	if fragmentFieldCount(fragment) == 0 {
		return fmt.Errorf("fragment must declare at least one owner array")
	}
	owner.Scopes = appendPresent(owner.Scopes, fragment.Scopes)
	owner.MigrationDirs = appendPresent(owner.MigrationDirs, fragment.MigrationDirs)
	owner.StaticSQLDirs = appendPresent(owner.StaticSQLDirs, fragment.StaticSQLDirs)
	owner.PlatformSchemas = appendPresent(owner.PlatformSchemas, fragment.PlatformSchemas)
	owner.SchemaUsageGrants = appendPresent(owner.SchemaUsageGrants, fragment.SchemaUsageGrants)
	owner.PlatformRoutines = appendPresent(owner.PlatformRoutines, fragment.PlatformRoutines)
	owner.RoutineExecGrants = appendPresent(owner.RoutineExecGrants, fragment.RoutineExecGrants)
	owner.PlatformTables = appendPresent(owner.PlatformTables, fragment.PlatformTables)
	owner.ExternalAccesses = appendPresent(owner.ExternalAccesses, fragment.ExternalAccesses)
	owner.ForeignKeys = appendPresent(owner.ForeignKeys, fragment.ForeignKeys)
	owner.LegacyTables = appendPresent(owner.LegacyTables, fragment.LegacyTables)
	owner.DynamicNamespaces = appendPresent(owner.DynamicNamespaces, fragment.DynamicNamespaces)
	return nil
}

func appendPresent[T any](current, addition []T) []T {
	if addition == nil {
		return current
	}
	if current == nil {
		current = make([]T, 0, len(addition))
	}
	return append(current, addition...)
}

func fragmentFieldCount(fragment ownerFragmentDocument) int {
	fields := []bool{
		fragment.Scopes != nil, fragment.MigrationDirs != nil,
		fragment.StaticSQLDirs != nil, fragment.PlatformSchemas != nil,
		fragment.SchemaUsageGrants != nil, fragment.PlatformRoutines != nil,
		fragment.RoutineExecGrants != nil, fragment.PlatformTables != nil,
		fragment.ExternalAccesses != nil, fragment.ForeignKeys != nil,
		fragment.LegacyTables != nil, fragment.DynamicNamespaces != nil,
	}
	count := 0
	for _, present := range fields {
		if present {
			count++
		}
	}
	return count
}
