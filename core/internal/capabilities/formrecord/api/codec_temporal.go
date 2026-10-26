package recordkernel

import (
	"encoding/json"
	"time"
)

const timestampLayout = "2006-01-02T15:04:05.000000Z"
const timeLayout = "15:04:05.000000"

func decodeTemporal(typeID TypeID, raw []byte) (Value, error) {
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return Value{}, ErrInputInvalid
	}
	layout := time.DateOnly
	if typeID == TypeTimestamp {
		layout = timestampLayout
	} else if typeID == TypeTime {
		layout = timeLayout
	}
	parsed, err := time.Parse(layout, text)
	if err != nil || parsed.Format(layout) != text ||
		(typeID == TypeDate || typeID == TypeTimestamp) &&
			(parsed.Year() < 1 || parsed.Year() > 9999) {
		return Value{}, ErrInputInvalid
	}
	encoded, _ := json.Marshal(text)
	return typedValue(typeID, encoded), nil
}
