package api

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const maximumRequestSorts = 5

type RequestSort struct {
	field     string
	direction Direction
}

func NewRequestSort(field string, direction Direction) (RequestSort, error) {
	if !validRequestName(field) {
		return RequestSort{}, fmt.Errorf("REST Query sort field is invalid")
	}
	if !direction.valid() {
		return RequestSort{}, fmt.Errorf("REST Query sort direction is invalid")
	}
	return RequestSort{field: field, direction: direction}, nil
}

func (sort RequestSort) Field() string           { return sort.field }
func (sort RequestSort) Direction() Direction    { return sort.direction }
func (sort RequestSort) String() string          { return queryValueRedaction }
func (sort RequestSort) GoString() string        { return queryValueRedaction }
func (RequestSort) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
func (sort RequestSort) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}

func (sort RequestSort) valid() bool {
	return validRequestName(sort.field) && sort.direction.valid()
}

func cloneRequestSorts(values []RequestSort) ([]RequestSort, error) {
	if len(values) > maximumRequestSorts {
		return nil, fmt.Errorf("REST Query sort count exceeds its limit")
	}
	result := make([]RequestSort, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		if !value.valid() {
			return nil, fmt.Errorf("REST Query sort %d is invalid", index)
		}
		if _, duplicate := seen[value.field]; duplicate {
			return nil, fmt.Errorf("REST Query sort field is duplicated")
		}
		seen[value.field] = struct{}{}
		result[index] = value
	}
	return result, nil
}

func validRequestName(value string) bool {
	return value != "" && value != "_included" && utf8.ValidString(value) &&
		!strings.ContainsRune(value, '\x00') && norm.NFC.IsNormalString(value)
}
