package invoice

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
)

type testEffects struct {
	events []invoiceapp.FinalizedEventIntent
}

func (effects *testEffects) PublishFinalized(
	_ context.Context,
	event invoiceapp.FinalizedEventIntent,
	_ invoiceapp.Finalization,
) error {
	effects.events = append(effects.events, event)
	return nil
}

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
		target := reflect.ValueOf(destinations[index]).Elem()
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

func testStore(rows ...scriptedRow) *store {
	index := 0
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
		execute:  func(context.Context, string, ...any) error { return nil },
		allowed:  func(context.Context) bool { return true },
		notFound: func(error) bool { return false },
		effects:  &testEffects{},
	}
}

func testV7() uuid.UUID { return uuid.Must(uuid.NewV7()) }

type noRows struct{}

func (noRows) Error() string { return "no rows" }
