package compiler

import "crypto/sha256"

func predecessorStatementSeal(
	statement PredecessorStatement,
) ([sha256.Size]byte, error) {
	writer := newSealWriter("forgecrud.restquery.postgres.v1-predecessor-statement.v1")
	writer.writeString(statement.sql)
	writer.writeInt(statement.predecessorIndex)
	writer.writeInt(int64(statement.class))
	if writeStatementArguments(writer, statement.arguments) != nil ||
		writeStatementColumns(writer, statement.columns) != nil ||
		writeStatementBinding(writer, statement.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}
