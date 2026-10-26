package compiler

import (
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const maximumResultOutputKeyBytes = 256

type ResultColumn struct {
	fieldID              uuid.UUID
	fieldType            api.FieldType
	outputKey            string
	hasOutput            bool
	sortOrdinal          int
	sortDirection        api.Direction
	cursorCanonicalBytes int
	validated            bool
}

func (column ResultColumn) Validate() error {
	if !column.validated || column.fieldID == uuid.Nil ||
		column.fieldID.Variant() != uuid.RFC4122 || column.fieldType.Validate() != nil ||
		column.hasOutput != (column.outputKey != "") ||
		column.hasOutput && (column.outputKey == "_included" ||
			len(column.outputKey) > maximumResultOutputKeyBytes ||
			!utf8.ValidString(column.outputKey) ||
			strings.ContainsRune(column.outputKey, '\x00') ||
			!norm.NFC.IsNormalString(column.outputKey)) ||
		column.sortOrdinal < -1 || column.sortOrdinal > 5 || !column.validSortMetadata() {
		return ErrContract
	}
	return nil
}

func (column ResultColumn) validSortMetadata() bool {
	sorted := column.sortOrdinal >= 0
	if !sorted {
		return column.sortDirection == 0 && column.cursorCanonicalBytes == 0
	}
	return column.sortDirection.String() != "" &&
		column.cursorCanonicalBytes >= 1 && column.cursorCanonicalBytes <= 2048
}

func (column ResultColumn) FieldID() (uuid.UUID, error) {
	if err := column.Validate(); err != nil {
		return uuid.Nil, err
	}
	return column.fieldID, nil
}

func (column ResultColumn) LogicalType() (api.LogicalType, error) {
	if err := column.Validate(); err != nil {
		return 0, err
	}
	kind, err := column.fieldType.Kind()
	if err != nil {
		return 0, ErrContract
	}
	return kind, nil
}

func (column ResultColumn) FieldType() (api.FieldType, error) {
	if err := column.Validate(); err != nil {
		return api.FieldType{}, err
	}
	return column.fieldType, nil
}

func (column ResultColumn) OutputKey() (string, bool, error) {
	if err := column.Validate(); err != nil {
		return "", false, err
	}
	return column.outputKey, column.hasOutput, nil
}

func (column ResultColumn) SortOrdinal() (int, bool, error) {
	if err := column.Validate(); err != nil {
		return 0, false, err
	}
	return column.sortOrdinal, column.sortOrdinal >= 0, nil
}

func (column ResultColumn) SortMetadata() (api.Direction, int, bool, error) {
	if err := column.Validate(); err != nil {
		return 0, 0, false, err
	}
	present := column.sortOrdinal >= 0
	return column.sortDirection, column.cursorCanonicalBytes, present, nil
}

func (ResultColumn) String() string                        { return compilerRedaction }
func (ResultColumn) GoString() string                      { return compilerRedaction }
func (ResultColumn) LogValue() slog.Value                  { return slog.StringValue(compilerRedaction) }
func (column ResultColumn) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (ResultColumn) MarshalJSON() ([]byte, error)          { return rejectJSON() }
