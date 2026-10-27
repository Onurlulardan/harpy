package compiler

import "crypto/sha256"

func includeTargetStatementSeal(
	statement IncludeTargetStatement,
) ([sha256.Size]byte, error) {
	writer := newSealWriter("harpy.restquery.postgres.include-target-statement.v1")
	writer.writeString(statement.sql)
	writer.writeInt(int64(statement.stableIndex))
	writer.writeBytes(statement.stableFieldID[:])
	writer.writeInt(int64(statement.class))
	if writeStatementArguments(writer, statement.arguments) != nil ||
		writeStatementColumns(writer, statement.columns) != nil ||
		writeCatalogReference(writer, statement.catalog) != nil ||
		writeIncludeStatementBinding(writer, statement.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}
