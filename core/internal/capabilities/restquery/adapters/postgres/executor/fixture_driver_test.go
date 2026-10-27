package executor

import (
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func consumedFixture(t *testing.T, fixture batchFixture) consumedBatch {
	t.Helper()
	batch, err := consumeBatch(fixture.batch, fixture.plan)
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func readyDriver(
	t *testing.T,
	batch consumedBatch,
	rows [][]any,
	total int64,
) (*fakeDriver, []statement) {
	t.Helper()
	statements, err := statementsFor(batch)
	if err != nil {
		t.Fatal(err)
	}
	driver := &fakeDriver{
		owner: "forms", rows: map[string]*fakeRows{
			statements[0].sql: {values: rows},
		},
		rowValues: map[string]any{
			statements[0].explainSQL: explainDocument("1"),
		},
		rowErrors: map[string]error{}, queryErrors: map[string]error{},
	}
	if batch.hasExactCount {
		driver.rowValues[statements[1].explainSQL] = explainDocument("1")
		driver.rowValues[statements[1].sql] = total
	}
	return driver, statements
}

func fixtureRecordRow(
	t *testing.T,
	batch consumedBatch,
	name string,
	jsonValue any,
	id string,
) []any {
	t.Helper()
	columns, err := batch.record.Columns()
	if err != nil {
		t.Fatal(err)
	}
	values := make([]any, len(columns))
	for index, column := range columns {
		kind, err := column.LogicalType()
		if err != nil {
			t.Fatal(err)
		}
		switch kind {
		case api.LogicalTypeText:
			values[index] = name
		case api.LogicalTypeJSON:
			values[index] = jsonValue
		case api.LogicalTypeUUID:
			values[index] = [16]byte(uuid.MustParse(id))
		default:
			t.Fatalf("unexpected fixture column kind %s", kind)
		}
	}
	return values
}

func explainDocument(cost string) []byte {
	return []byte(fmt.Sprintf(`[{"Plan":{"Total Cost":%s}}]`, cost))
}

func eventIndex(events []string, target string) int {
	for index, event := range events {
		if event == target {
			return index
		}
	}
	return -1
}
