package recordpage

import (
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

type statementMetadata struct {
	columns       []columnMetadata
	sorts         []sortMetadata
	context       cursor.Context
	requestedSize int
}

type columnMetadata struct {
	fieldID     uuid.UUID
	fieldType   api.FieldType
	kind        api.LogicalType
	outputKey   string
	hasOutput   bool
	sortOrdinal int
	sorted      bool
}

type sortMetadata struct {
	fieldID        uuid.UUID
	direction      api.Direction
	kind           api.LogicalType
	canonicalBound int
}

func loadMetadata(statement compiler.RecordStatement) (statementMetadata, error) {
	class, classErr := statement.Class()
	columns, columnErr := statement.Columns()
	requested, sizeErr := statement.RequestedSize()
	context, contextErr := statement.CursorContext()
	pageSize, pageErr := context.PageSize()
	sorts, sortsErr := context.Sorts()
	if classErr != nil || columnErr != nil || sizeErr != nil || contextErr != nil ||
		pageErr != nil || sortsErr != nil || class != planner.StatementClassRecord ||
		requested != pageSize || len(columns) < 1 || len(sorts) < 1 {
		return statementMetadata{}, ErrContract
	}
	metadata := statementMetadata{
		columns: make([]columnMetadata, len(columns)), context: context,
		sorts: make([]sortMetadata, len(sorts)), requestedSize: requested,
	}
	for index, column := range columns {
		loaded, err := loadColumn(column)
		if err != nil {
			return statementMetadata{}, ErrContract
		}
		metadata.columns[index] = loaded
	}
	for index, sort := range sorts {
		loaded, err := loadSort(sort)
		if err != nil {
			return statementMetadata{}, ErrContract
		}
		metadata.sorts[index] = loaded
	}
	if validateMetadata(metadata, sorts) != nil {
		return statementMetadata{}, ErrContract
	}
	return metadata, nil
}

func loadSort(sort cursor.SortDefinition) (sortMetadata, error) {
	fieldID, idErr := sort.FieldID()
	direction, directionErr := sort.Direction()
	kind, typeErr := sort.LogicalType()
	bound, boundErr := sort.CanonicalBound()
	if idErr != nil || directionErr != nil || typeErr != nil || boundErr != nil {
		return sortMetadata{}, ErrContract
	}
	return sortMetadata{
		fieldID: fieldID, direction: direction, kind: kind, canonicalBound: bound,
	}, nil
}

func loadColumn(column compiler.ResultColumn) (columnMetadata, error) {
	fieldID, idErr := column.FieldID()
	fieldType, typeErr := column.FieldType()
	kind, kindErr := column.LogicalType()
	key, output, keyErr := column.OutputKey()
	ordinal, sorted, ordinalErr := column.SortOrdinal()
	if idErr != nil || typeErr != nil || kindErr != nil || keyErr != nil || ordinalErr != nil {
		return columnMetadata{}, ErrContract
	}
	return columnMetadata{
		fieldID: fieldID, fieldType: fieldType, kind: kind,
		outputKey: key, hasOutput: output, sortOrdinal: ordinal, sorted: sorted,
	}, nil
}

func validOutputKey(key string) bool {
	return key != "" && key != "_included" && utf8.ValidString(key) &&
		!strings.ContainsRune(key, '\x00') && norm.NFC.IsNormalString(key)
}
