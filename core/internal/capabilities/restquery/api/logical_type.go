package api

import (
	"errors"
	"fmt"
	"io"
)

var errInvalidLogicalType = errors.New("REST Query logical type is invalid")

type LogicalType uint8

const (
	LogicalTypeText LogicalType = iota + 1
	LogicalTypeInteger
	LogicalTypeDecimal
	LogicalTypeBoolean
	LogicalTypeDate
	LogicalTypeTimestamp
	LogicalTypeUUID
	LogicalTypeEnum
	LogicalTypeRelation
	LogicalTypeTextList
	LogicalTypeUUIDList
	LogicalTypeJSON
)

func ParseLogicalType(value string) (LogicalType, bool) {
	switch value {
	case "text":
		return LogicalTypeText, true
	case "integer":
		return LogicalTypeInteger, true
	case "decimal":
		return LogicalTypeDecimal, true
	case "boolean":
		return LogicalTypeBoolean, true
	case "date":
		return LogicalTypeDate, true
	case "timestamp":
		return LogicalTypeTimestamp, true
	case "uuid":
		return LogicalTypeUUID, true
	case "enum":
		return LogicalTypeEnum, true
	case "relation":
		return LogicalTypeRelation, true
	case "text_list":
		return LogicalTypeTextList, true
	case "uuid_list":
		return LogicalTypeUUIDList, true
	case "json":
		return LogicalTypeJSON, true
	default:
		return 0, false
	}
}

func (logicalType LogicalType) String() string {
	switch logicalType {
	case LogicalTypeText:
		return "text"
	case LogicalTypeInteger:
		return "integer"
	case LogicalTypeDecimal:
		return "decimal"
	case LogicalTypeBoolean:
		return "boolean"
	case LogicalTypeDate:
		return "date"
	case LogicalTypeTimestamp:
		return "timestamp"
	case LogicalTypeUUID:
		return "uuid"
	case LogicalTypeEnum:
		return "enum"
	case LogicalTypeRelation:
		return "relation"
	case LogicalTypeTextList:
		return "text_list"
	case LogicalTypeUUIDList:
		return "uuid_list"
	case LogicalTypeJSON:
		return "json"
	default:
		return ""
	}
}

func (logicalType LogicalType) GoString() string { return logicalType.String() }

func (logicalType LogicalType) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, logicalType.String())
}

func (logicalType LogicalType) Valid() bool {
	return logicalType >= LogicalTypeText && logicalType <= LogicalTypeJSON
}

func (logicalType LogicalType) Validate() error {
	if !logicalType.Valid() {
		return errInvalidLogicalType
	}
	return nil
}

func (LogicalType) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
