package postgres

func catalogMatchesStartup(catalog catalogPreflight, startup StartupResult) bool {
	origin, err := originForVectorSchema(catalog.vectorSchema)
	return err == nil && catalog.serverVersion == startup.serverVersion &&
		uint32(catalog.databaseOID) == startup.databaseOID &&
		catalog.databaseName == startup.databaseName &&
		catalog.systemIdentifier == startup.systemIdentifier &&
		catalog.vectorSchema == startup.vectorSchema && origin == startup.databaseOrigin
}
