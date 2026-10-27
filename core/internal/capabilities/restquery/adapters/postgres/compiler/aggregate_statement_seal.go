package compiler

import "crypto/sha256"

func aggregateStatementSeal(
	statement AggregateStatement,
) ([sha256.Size]byte, error) {
	writer := newSealWriter("harpy.restquery.postgres.aggregate-statement.v1")
	writer.writeString(statement.sql)
	writer.writeInt(int64(statement.limit))
	writer.writeInt(int64(statement.class))
	if writeStatementArguments(writer, statement.arguments) != nil ||
		writeAggregateGroups(writer, statement.groups) != nil ||
		writeAggregateMeasures(writer, statement.measures) != nil ||
		writeStatementBinding(writer, statement.binding) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	return writer.sum()
}

func writeAggregateGroups(
	writer *sealWriter, columns []AggregateGroupColumn,
) error {
	writer.writeInt(int64(len(columns)))
	for _, column := range columns {
		if column.Validate() != nil || writeFieldType(writer, column.fieldType) != nil {
			return ErrContract
		}
		writer.writeBytes(column.fieldID[:])
		writer.writeString(column.outputKey)
	}
	return nil
}

func writeAggregateMeasures(
	writer *sealWriter, columns []AggregateMeasureColumn,
) error {
	writer.writeInt(int64(len(columns)))
	for _, column := range columns {
		if column.Validate() != nil || writeFieldType(writer, column.fieldType) != nil {
			return ErrContract
		}
		writer.writeBytes(column.sourceID[:])
		writer.writeBool(column.hasSource)
		writer.writeString(column.outputKey)
		writer.writeInt(int64(column.function))
	}
	return nil
}
