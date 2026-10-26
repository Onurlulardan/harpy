package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
)

// PredecessorTuple is a sealed typed tuple bound to one predecessor statement.
type PredecessorTuple struct {
	values          []cursor.LastValue
	columns         []ResultColumn
	binding         statementBinding
	predecessorSeal [sha256.Size]byte
	stateSeal       [sha256.Size]byte
	predecessor     int64
	validated       bool
}

func (statement PredecessorStatement) MintTuple(
	values []cursor.LastValue,
) (PredecessorTuple, error) {
	if statement.Validate() != nil {
		return PredecessorTuple{}, ErrContract
	}
	tuple := PredecessorTuple{
		values:  append([]cursor.LastValue{}, values...),
		columns: append([]ResultColumn{}, statement.columns...),
		binding: statement.binding, predecessorSeal: statement.stateSeal,
		predecessor: statement.predecessorIndex, validated: true,
	}
	seal, err := predecessorTupleSeal(tuple)
	if err != nil {
		return PredecessorTuple{}, ErrContract
	}
	tuple.stateSeal = seal
	if tuple.Validate() != nil {
		return PredecessorTuple{}, ErrContract
	}
	return tuple, nil
}

func (tuple PredecessorTuple) Validate() error {
	if !tuple.validated || tuple.binding.validate() != nil || tuple.predecessor < 0 ||
		tuple.predecessorSeal == [sha256.Size]byte{} ||
		tuple.stateSeal == [sha256.Size]byte{} ||
		validateTupleValues(tuple.columns, tuple.values) != nil {
		return ErrContract
	}
	seal, err := predecessorTupleSeal(tuple)
	if err != nil || seal != tuple.stateSeal {
		return ErrContract
	}
	return nil
}

func (tuple PredecessorTuple) validateFor(statement PredecessorStatement) error {
	if tuple.Validate() != nil || statement.Validate() != nil ||
		tuple.binding != statement.binding || tuple.predecessor != statement.predecessorIndex ||
		tuple.predecessorSeal != statement.stateSeal {
		return ErrContract
	}
	return nil
}

func (PredecessorTuple) String() string   { return compilerRedaction }
func (PredecessorTuple) GoString() string { return compilerRedaction }
func (PredecessorTuple) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (tuple PredecessorTuple) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (PredecessorTuple) MarshalJSON() ([]byte, error)         { return rejectJSON() }
