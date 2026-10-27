package systemresource

import (
	"fmt"
	"log/slog"
)

const redactedValue = "[REDACTED]"

func writeRedacted(state fmt.State) {
	_, _ = state.Write([]byte(redactedValue))
}

func redactedLogValue() slog.Value {
	return slog.StringValue(redactedValue)
}

func rejectDirectJSON() ([]byte, error) {
	return nil, ErrContract
}

func (ActorEvidence) String() string                 { return redactedValue }
func (ActorEvidence) GoString() string               { return redactedValue }
func (ActorEvidence) LogValue() slog.Value           { return redactedLogValue() }
func (ActorEvidence) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (ActorEvidence) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (FieldValues) String() string                 { return redactedValue }
func (FieldValues) GoString() string               { return redactedValue }
func (FieldValues) LogValue() slog.Value           { return redactedLogValue() }
func (FieldValues) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (FieldValues) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (CreateInput) String() string                 { return redactedValue }
func (CreateInput) GoString() string               { return redactedValue }
func (CreateInput) LogValue() slog.Value           { return redactedLogValue() }
func (CreateInput) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (CreateInput) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (GetInput) String() string                 { return redactedValue }
func (GetInput) GoString() string               { return redactedValue }
func (GetInput) LogValue() slog.Value           { return redactedLogValue() }
func (GetInput) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (GetInput) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (UpdateInput) String() string                 { return redactedValue }
func (UpdateInput) GoString() string               { return redactedValue }
func (UpdateInput) LogValue() slog.Value           { return redactedLogValue() }
func (UpdateInput) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (UpdateInput) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (DeleteInput) String() string                 { return redactedValue }
func (DeleteInput) GoString() string               { return redactedValue }
func (DeleteInput) LogValue() slog.Value           { return redactedLogValue() }
func (DeleteInput) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (DeleteInput) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (StateInput) String() string                 { return redactedValue }
func (StateInput) GoString() string               { return redactedValue }
func (StateInput) LogValue() slog.Value           { return redactedLogValue() }
func (StateInput) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (StateInput) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (QueryInput) String() string                 { return redactedValue }
func (QueryInput) GoString() string               { return redactedValue }
func (QueryInput) LogValue() slog.Value           { return redactedLogValue() }
func (QueryInput) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (QueryInput) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (Record) String() string                 { return redactedValue }
func (Record) GoString() string               { return redactedValue }
func (Record) LogValue() slog.Value           { return redactedLogValue() }
func (Record) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (Record) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (MutationReceipt) String() string                 { return redactedValue }
func (MutationReceipt) GoString() string               { return redactedValue }
func (MutationReceipt) LogValue() slog.Value           { return redactedLogValue() }
func (MutationReceipt) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (MutationReceipt) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (FieldCatalog) String() string                 { return redactedValue }
func (FieldCatalog) GoString() string               { return redactedValue }
func (FieldCatalog) LogValue() slog.Value           { return redactedLogValue() }
func (FieldCatalog) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (FieldCatalog) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (FieldDescriptor) String() string                 { return redactedValue }
func (FieldDescriptor) GoString() string               { return redactedValue }
func (FieldDescriptor) LogValue() slog.Value           { return redactedLogValue() }
func (FieldDescriptor) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (FieldDescriptor) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (ReadResult) String() string                 { return redactedValue }
func (ReadResult) GoString() string               { return redactedValue }
func (ReadResult) LogValue() slog.Value           { return redactedLogValue() }
func (ReadResult) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (ReadResult) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (MutationResult) String() string                 { return redactedValue }
func (MutationResult) GoString() string               { return redactedValue }
func (MutationResult) LogValue() slog.Value           { return redactedLogValue() }
func (MutationResult) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (MutationResult) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (DeleteResult) String() string                 { return redactedValue }
func (DeleteResult) GoString() string               { return redactedValue }
func (DeleteResult) LogValue() slog.Value           { return redactedLogValue() }
func (DeleteResult) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (DeleteResult) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (QueryResult) String() string                 { return redactedValue }
func (QueryResult) GoString() string               { return redactedValue }
func (QueryResult) LogValue() slog.Value           { return redactedLogValue() }
func (QueryResult) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (QueryResult) Format(state fmt.State, _ rune) { writeRedacted(state) }
