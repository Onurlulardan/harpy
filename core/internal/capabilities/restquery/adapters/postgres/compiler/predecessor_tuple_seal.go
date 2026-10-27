package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
)

func predecessorTupleSeal(tuple PredecessorTuple) ([sha256.Size]byte, error) {
	writer := newSealWriter("harpy.restquery.postgres.v1-predecessor-tuple.v1")
	writer.writeInt(tuple.predecessor)
	writer.writeBytes(tuple.predecessorSeal[:])
	if writeStatementBinding(writer, tuple.binding) != nil ||
		writeStatementColumns(writer, tuple.columns) != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	writer.writeInt(int64(len(tuple.values)))
	for _, value := range tuple.values {
		if writeTupleLastValue(writer, value) != nil {
			return [sha256.Size]byte{}, ErrContract
		}
	}
	return writer.sum()
}

func writeTupleLastValue(writer *sealWriter, value cursor.LastValue) error {
	null, err := value.IsNull()
	if err != nil {
		return ErrContract
	}
	writer.writeBool(null)
	if null {
		return nil
	}
	resolvedValue, present, err := value.Value()
	if err != nil || !present {
		return ErrContract
	}
	kind, err := resolvedValue.Kind()
	if err != nil {
		return ErrContract
	}
	canonical, err := tupleCanonicalBytes(resolvedValue, kind)
	if err != nil || kind < api.LogicalTypeText || kind > api.LogicalTypeRelation {
		return ErrContract
	}
	writer.writeInt(int64(kind))
	writer.writeBytes(canonical)
	return nil
}
