package api

import (
	"strconv"
	"time"
)

type ValueKind string

const (
	ValueString  ValueKind = "string"
	ValueInteger ValueKind = "integer"
	ValueBoolean ValueKind = "boolean"
	ValueTime    ValueKind = "time"
)

type TemplateValue struct {
	Key     string
	Kind    ValueKind
	String  string
	Integer int64
	Boolean bool
	Time    time.Time
}

func (value TemplateValue) Render() (string, bool) {
	switch value.Kind {
	case ValueString:
		return value.String, value.String != "" && value.Integer == 0 &&
			!value.Boolean && value.Time.IsZero()
	case ValueInteger:
		return strconv.FormatInt(value.Integer, 10), value.String == "" &&
			!value.Boolean && value.Time.IsZero()
	case ValueBoolean:
		return strconv.FormatBool(value.Boolean), value.String == "" &&
			value.Integer == 0 && value.Time.IsZero()
	case ValueTime:
		return value.Time.UTC().Format(time.RFC3339Nano), value.String == "" &&
			value.Integer == 0 && !value.Boolean && !value.Time.IsZero()
	default:
		return "", false
	}
}
