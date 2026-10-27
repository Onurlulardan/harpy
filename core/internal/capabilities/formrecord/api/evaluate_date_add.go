package recordkernel

import (
	"encoding/json"
	"time"
)

const maximumCalendarDays int64 = 3_652_058

func evaluateDateAdd(typeID TypeID, value, offset Value) (Value, error) {
	text, textErr := valueString(value)
	days, daysErr := valueInteger(offset)
	if textErr != nil || daysErr != nil || days < -maximumCalendarDays ||
		days > maximumCalendarDays {
		return Value{}, ErrCalculation
	}
	layout := time.DateOnly
	if typeID == TypeTimestamp {
		layout = timestampLayout
	}
	parsed, err := time.Parse(layout, text)
	if err != nil {
		return Value{}, ErrCalculation
	}
	result := parsed.AddDate(0, 0, int(days))
	if result.Year() < 1 || result.Year() > 9999 {
		return Value{}, ErrCalculation
	}
	encoded, _ := json.Marshal(result.Format(layout))
	return typedValue(typeID, encoded), nil
}
