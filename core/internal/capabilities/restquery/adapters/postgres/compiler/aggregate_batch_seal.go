package compiler

import "crypto/sha256"

func aggregateBatchSeal(batch AggregateBatch) ([sha256.Size]byte, error) {
	writer := newSealWriter("forgecrud.restquery.postgres.aggregate-batch.v1")
	writer.writeBytes(batch.statement.stateSeal[:])
	if writeStatementBinding(writer, batch.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}
