package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type RecordBatch struct {
	record        RecordStatement
	exactCount    ExactCountStatement
	countMode     api.CountMode
	binding       statementBinding
	stateSeal     [sha256.Size]byte
	hasExactCount bool
	validated     bool
}

func newRecordBatch(
	record RecordStatement,
	exactCount ExactCountStatement,
	countMode api.CountMode,
	binding statementBinding,
) (RecordBatch, error) {
	batch := RecordBatch{
		record: record, exactCount: exactCount, countMode: countMode,
		binding: binding, hasExactCount: countMode == api.CountModeExact,
		validated: true,
	}
	seal, err := recordBatchSeal(batch)
	if err != nil {
		return RecordBatch{}, ErrContract
	}
	batch.stateSeal = seal
	if err := batch.Validate(); err != nil {
		return RecordBatch{}, err
	}
	return batch, nil
}

func (batch RecordBatch) Validate() error {
	if !batch.validated || batch.record.Validate() != nil ||
		batch.binding.validate() != nil || batch.record.binding != batch.binding ||
		batch.record.countMode != batch.countMode ||
		batch.stateSeal == [sha256.Size]byte{} || !batch.validCountShape() {
		return ErrContract
	}
	seal, err := recordBatchSeal(batch)
	if err != nil || seal != batch.stateSeal {
		return ErrContract
	}
	return nil
}

func (batch RecordBatch) validCountShape() bool {
	switch batch.countMode {
	case api.CountModeNone:
		return !batch.hasExactCount && exactCountStatementIsZero(batch.exactCount)
	case api.CountModeExact:
		return batch.hasExactCount && batch.exactCount.Validate() == nil &&
			batch.exactCount.binding == batch.binding
	default:
		return false
	}
}

func exactCountStatementIsZero(statement ExactCountStatement) bool {
	return statement.sql == "" && statement.arguments == nil &&
		statement.class == 0 && statement.binding == (statementBinding{}) &&
		statement.stateSeal == [sha256.Size]byte{} && !statement.validated
}

func (RecordBatch) String() string   { return compilerRedaction }
func (RecordBatch) GoString() string { return compilerRedaction }
func (RecordBatch) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (batch RecordBatch) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (RecordBatch) MarshalJSON() ([]byte, error)         { return rejectJSON() }
