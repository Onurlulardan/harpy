package compiler

import "crypto/sha256"

func recordBatchSeal(batch RecordBatch) ([sha256.Size]byte, error) {
	writer := newSealWriter("harpy.restquery.postgres.record-batch.v1")
	writer.writeInt(int64(batch.countMode))
	writer.writeBool(batch.hasExactCount)
	writer.writeBytes(batch.record.stateSeal[:])
	writer.writeBytes(batch.exactCount.stateSeal[:])
	if writeStatementBinding(writer, batch.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}
