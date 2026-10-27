package executor

import (
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/google/uuid"
)

type uuidValue = uuid.UUID

func materialRootPage(
	t *testing.T,
	identifiers []uuid.UUID,
) (recordpage.Page, []materialRecord) {
	t.Helper()
	fixture := newBatchFixtureWithPageSize(
		t, api.CountModeNone, fixtureTenant, fixtureActor, max(1, len(identifiers)),
	)
	batch := consumedFixture(t, fixture)
	rows := make([][]any, len(identifiers))
	for index, identifier := range identifiers {
		rows[index] = fixtureRecordRow(
			t, batch, fmt.Sprintf("root-%d", index), nil, identifier.String(),
		)
	}
	driver, _ := readyDriver(t, batch, rows, 0)
	result, err := runBatch(
		t.Context(), driver, batch, fixture.codec, planner.HardPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	page, err := result.Page()
	if err != nil {
		t.Fatal(err)
	}
	documents, err := page.Records()
	if err != nil {
		t.Fatal(err)
	}
	records := make([]materialRecord, len(identifiers))
	for index, identifier := range identifiers {
		records[index] = materialFixtureRecord(t, identifier, documents[index])
	}
	return page, records
}

func materialFixtureRecord(
	t *testing.T,
	identifier uuid.UUID,
	document []byte,
) materialRecord {
	t.Helper()
	record, err := newMaterialRecord(
		identifier, document, [][]byte{[]byte(`"` + identifier.String() + `"`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func materialUUID(value int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("40000000-0000-4000-8000-%012d", value))
}

func materialTargets(values ...uuid.UUID) map[uuid.UUID]struct{} {
	result := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
