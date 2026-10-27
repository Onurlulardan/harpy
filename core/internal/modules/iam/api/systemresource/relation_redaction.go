package systemresource

import (
	"fmt"
	"log/slog"
)

func (RelationSummary) String() string                 { return redactedValue }
func (RelationSummary) GoString() string               { return redactedValue }
func (RelationSummary) LogValue() slog.Value           { return redactedLogValue() }
func (RelationSummary) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (RelationSummary) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (RelationValue) String() string                 { return redactedValue }
func (RelationValue) GoString() string               { return redactedValue }
func (RelationValue) LogValue() slog.Value           { return redactedLogValue() }
func (RelationValue) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (RelationValue) Format(state fmt.State, _ rune) { writeRedacted(state) }

func (RelationProjection) String() string                 { return redactedValue }
func (RelationProjection) GoString() string               { return redactedValue }
func (RelationProjection) LogValue() slog.Value           { return redactedLogValue() }
func (RelationProjection) MarshalJSON() ([]byte, error)   { return rejectDirectJSON() }
func (RelationProjection) Format(state fmt.State, _ rune) { writeRedacted(state) }
