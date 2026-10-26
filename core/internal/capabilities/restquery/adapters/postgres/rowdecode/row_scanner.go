package rowdecode

import (
	"bytes"
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

const maximumRowColumns = 56
const pendingJSONMarker = "\xff\x00"

type pendingDriverValue struct{}

type scannerState uint8

const (
	scannerReady scannerState = iota + 1
	scannerIssued
	scannerConsumed
)

type scanSlot struct {
	driverValue any
	jsonValue   []byte
	json        bool
}

// RowScanner owns one row's exact scan targets and can be consumed only once.
type RowScanner struct {
	state *rowScannerState
}

type rowScannerState struct {
	mu         sync.Mutex
	fieldTypes []api.FieldType
	slots      []scanSlot
	phase      scannerState
}

func NewRowScanner(fieldTypes []api.FieldType) (*RowScanner, error) {
	if len(fieldTypes) < 1 || len(fieldTypes) > maximumRowColumns {
		return nil, ErrInvalidRow
	}
	state := &rowScannerState{
		fieldTypes: append([]api.FieldType{}, fieldTypes...),
		slots:      make([]scanSlot, len(fieldTypes)),
		phase:      scannerReady,
	}
	for index, fieldType := range state.fieldTypes {
		kind, err := fieldType.Kind()
		if err != nil {
			return nil, ErrInvalidFieldType
		}
		if kind == api.LogicalTypeJSON {
			state.slots[index] = scanSlot{
				jsonValue: []byte(pendingJSONMarker), json: true,
			}
			continue
		}
		state.slots[index].driverValue = pendingDriverValue{}
	}
	return &RowScanner{state: state}, nil
}

// Destinations returns exact pgx scan targets once. JSON always uses raw bytes.
func (scanner *RowScanner) Destinations() ([]any, error) {
	if scanner == nil || scanner.state == nil {
		return nil, ErrInvalidRow
	}
	state := scanner.state
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.phase != scannerReady {
		state.phase = scannerConsumed
		return nil, ErrInvalidRow
	}
	state.phase = scannerIssued
	destinations := make([]any, len(state.slots))
	for index := range state.slots {
		if state.slots[index].json {
			destinations[index] = &state.slots[index].jsonValue
		} else {
			destinations[index] = &state.slots[index].driverValue
		}
	}
	return destinations, nil
}

// Decode consumes the scanned row once and never returns partial output.
func (scanner *RowScanner) Decode() ([]Value, error) {
	if scanner == nil || scanner.state == nil {
		return nil, ErrInvalidRow
	}
	state := scanner.state
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.phase != scannerIssued {
		state.phase = scannerConsumed
		return nil, ErrInvalidRow
	}
	state.phase = scannerConsumed
	driverValues := make([]any, len(state.slots))
	for index, slot := range state.slots {
		if slot.json {
			if string(slot.jsonValue) == pendingJSONMarker {
				return nil, ErrInvalidRow
			}
			if slot.jsonValue != nil {
				driverValues[index] = bytes.Clone(slot.jsonValue)
			}
			continue
		}
		if _, pending := slot.driverValue.(pendingDriverValue); pending {
			return nil, ErrInvalidRow
		}
		driverValues[index] = slot.driverValue
	}
	return DecodeRow(state.fieldTypes, driverValues)
}
