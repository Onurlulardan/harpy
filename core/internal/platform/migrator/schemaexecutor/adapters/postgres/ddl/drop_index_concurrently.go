package ddl

func DropIndexConcurrently(schemaName, indexName string) Statement {
	return Statement{Text: "DROP INDEX CONCURRENTLY " + qualify(schemaName, indexName)}
}
