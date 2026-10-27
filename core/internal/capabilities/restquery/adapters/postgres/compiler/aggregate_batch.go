package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
)

type AggregateBatch struct {
	statement AggregateStatement
	binding   statementBinding
	stateSeal [sha256.Size]byte
	validated bool
}

func newAggregateBatch(
	statement AggregateStatement, binding statementBinding,
) (AggregateBatch, error) {
	batch := AggregateBatch{
		statement: statement, binding: binding, validated: true,
	}
	seal, err := aggregateBatchSeal(batch)
	if err != nil {
		return AggregateBatch{}, ErrContract
	}
	batch.stateSeal = seal
	if batch.Validate() != nil {
		return AggregateBatch{}, ErrContract
	}
	return batch, nil
}

func (batch AggregateBatch) Validate() error {
	if !batch.validated || batch.statement.Validate() != nil ||
		batch.binding.validate() != nil || batch.statement.binding != batch.binding ||
		batch.stateSeal == [sha256.Size]byte{} {
		return ErrContract
	}
	seal, err := aggregateBatchSeal(batch)
	if err != nil || seal != batch.stateSeal {
		return ErrContract
	}
	return nil
}

func (AggregateBatch) String() string   { return compilerRedaction }
func (AggregateBatch) GoString() string { return compilerRedaction }
func (AggregateBatch) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (batch AggregateBatch) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (AggregateBatch) MarshalJSON() ([]byte, error)         { return rejectJSON() }
