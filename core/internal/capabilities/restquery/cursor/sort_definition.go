package cursor

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

type SortDefinition struct {
	fieldID        uuid.UUID
	direction      api.Direction
	logicalType    api.LogicalType
	canonicalBound int
	valid          bool
}

func NewSortDefinition(
	fieldID string,
	direction api.Direction,
	logicalType api.LogicalType,
	canonicalBound int,
) (SortDefinition, error) {
	identifier, parsed := parseCanonicalUUID(fieldID)
	sort := SortDefinition{
		fieldID: identifier, direction: direction, logicalType: logicalType,
		canonicalBound: canonicalBound, valid: parsed,
	}
	if err := sort.Validate(); err != nil {
		return SortDefinition{}, err
	}
	return sort, nil
}

func (sort SortDefinition) Validate() error {
	if !sort.valid || sort.fieldID == uuid.Nil || !validDirection(sort.direction) ||
		!sortableLogicalType(sort.logicalType) || sort.canonicalBound < 1 ||
		sort.canonicalBound > maximumCanonicalValueBytes {
		return errInvalidSort
	}
	return nil
}

func (sort SortDefinition) FieldID() (uuid.UUID, error) {
	if err := sort.Validate(); err != nil {
		return uuid.Nil, err
	}
	return sort.fieldID, nil
}

func (sort SortDefinition) Direction() (api.Direction, error) {
	if err := sort.Validate(); err != nil {
		return 0, err
	}
	return sort.direction, nil
}

func (sort SortDefinition) LogicalType() (api.LogicalType, error) {
	if err := sort.Validate(); err != nil {
		return 0, err
	}
	return sort.logicalType, nil
}

func (sort SortDefinition) CanonicalBound() (int, error) {
	if err := sort.Validate(); err != nil {
		return 0, err
	}
	return sort.canonicalBound, nil
}

func (SortDefinition) String() string                      { return redacted }
func (SortDefinition) GoString() string                    { return redacted }
func (SortDefinition) LogValue() slog.Value                { return redactedLogValue() }
func (sort SortDefinition) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (SortDefinition) MarshalJSON() ([]byte, error)        { return rejectJSONMarshal() }

func validDirection(direction api.Direction) bool {
	parsed, ok := api.ParseDirection(direction.String())
	return ok && parsed == direction
}

func sortableLogicalType(logicalType api.LogicalType) bool {
	switch logicalType {
	case api.LogicalTypeText, api.LogicalTypeInteger, api.LogicalTypeDecimal,
		api.LogicalTypeBoolean, api.LogicalTypeDate, api.LogicalTypeTimestamp,
		api.LogicalTypeUUID, api.LogicalTypeEnum, api.LogicalTypeRelation:
		return true
	default:
		return false
	}
}
