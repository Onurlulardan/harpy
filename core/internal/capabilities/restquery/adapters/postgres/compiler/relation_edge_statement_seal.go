package compiler

import "crypto/sha256"

func relationEdgeStatementSeal(
	statement RelationEdgeStatement,
) ([sha256.Size]byte, error) {
	writer := newSealWriter("harpy.restquery.postgres.relation-edge-statement.v1")
	writer.writeString(statement.sql)
	writer.writeInt(int64(statement.class))
	if writeStatementArguments(writer, statement.arguments) != nil ||
		writeCatalogReference(writer, statement.catalog) != nil ||
		writeIncludeStatementBinding(writer, statement.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}
