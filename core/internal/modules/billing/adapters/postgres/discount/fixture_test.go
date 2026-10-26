package discount

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

type scriptedRow struct {
	values []any
	err    error
}

func (script scriptedRow) Scan(destinations ...any) error {
	if script.err != nil {
		return script.err
	}
	if len(destinations) != len(script.values) {
		return fmt.Errorf("scan destinations=%d values=%d", len(destinations), len(script.values))
	}
	for index, value := range script.values {
		destination := reflect.ValueOf(destinations[index])
		if destination.Kind() != reflect.Pointer || destination.IsNil() {
			return fmt.Errorf("destination %d is not writable", index)
		}
		target := destination.Elem()
		if value == nil {
			target.Set(reflect.Zero(target.Type()))
			continue
		}
		source := reflect.ValueOf(value)
		if !source.Type().AssignableTo(target.Type()) {
			return fmt.Errorf("value %d type %s cannot assign to %s",
				index, source.Type(), target.Type())
		}
		target.Set(source)
	}
	return nil
}

func testStore(rows ...scriptedRow) (*store, *int) {
	index := 0
	executions := 0
	return &store{
		tenantID: uuid.MustParse("71000000-0000-4000-8000-000000000071"),
		queryRow: func(context.Context, string, ...any) row {
			if index >= len(rows) {
				return scriptedRow{err: fmt.Errorf("unexpected query")}
			}
			selected := rows[index]
			index++
			return selected
		},
		execute: func(context.Context, string, ...any) error {
			executions++
			return nil
		},
		allowed:  func(context.Context) bool { return true },
		notFound: func(error) bool { return false },
	}, &executions
}

func testV7() uuid.UUID { return uuid.Must(uuid.NewV7()) }
