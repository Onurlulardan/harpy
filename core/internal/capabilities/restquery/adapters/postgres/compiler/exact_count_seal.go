package compiler

import "crypto/sha256"

func exactCountStatementSeal(
	statement ExactCountStatement,
) ([sha256.Size]byte, error) {
	writer := newSealWriter(
		"harpy.restquery.postgres.exact-count-statement.v1",
	)
	writer.writeString(statement.sql)
	writer.writeInt(int64(statement.class))
	if writeStatementArguments(writer, statement.arguments) != nil ||
		writeStatementBinding(writer, statement.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}
