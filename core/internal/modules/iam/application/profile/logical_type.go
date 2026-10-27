package profile

type LogicalType string

const (
	LogicalString    LogicalType = "string"
	LogicalInteger   LogicalType = "integer"
	LogicalDecimal   LogicalType = "decimal"
	LogicalBoolean   LogicalType = "boolean"
	LogicalDate      LogicalType = "date"
	LogicalTimestamp LogicalType = "timestamp"
	LogicalUUID      LogicalType = "uuid"
	LogicalEnum      LogicalType = "enum"
	LogicalJSON      LogicalType = "json"
)

func (kind LogicalType) Valid() bool {
	switch kind {
	case LogicalString, LogicalInteger, LogicalDecimal, LogicalBoolean,
		LogicalDate, LogicalTimestamp, LogicalUUID, LogicalEnum, LogicalJSON:
		return true
	default:
		return false
	}
}
