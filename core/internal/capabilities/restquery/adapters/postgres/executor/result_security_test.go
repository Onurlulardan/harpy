package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestResultIsSealedRedactedAndDefensivelyCopied(t *testing.T) {
	fixture := newBatchFixture(
		t, api.CountModeExact, fixtureTenant, fixtureActor,
	)
	batch := consumedFixture(t, fixture)
	record := fixtureRecordRow(
		t, batch, "Ada", []byte(`{"safe":true}`),
		"40000000-0000-4000-8000-000000000001",
	)
	driver, _ := readyDriver(t, batch, [][]any{record}, 7)
	result, err := runBatch(
		context.Background(), driver, batch, fixture.codec, planner.HardPolicy(),
	)
	if err != nil {
		t.Fatal(err)
	}
	page, _ := result.Page()
	first, _ := page.Records()
	first[0][0] = 'X'
	second, _ := page.Records()
	if string(second[0]) != `{"name":"Ada","metadata":{"safe":true}}` {
		t.Fatalf("result records were mutated: %s", second[0])
	}
	for _, rendered := range []string{
		result.String(), result.GoString(), fmt.Sprintf("%v", result),
		fmt.Sprintf("%#v", result),
	} {
		if rendered != redacted || strings.Contains(rendered, "Ada") {
			t.Fatalf("result formatting leaked: %q", rendered)
		}
	}
	if _, err := json.Marshal(result); !errors.Is(err, ErrDirectMarshal) {
		t.Fatalf("direct marshal error=%v", err)
	}
	tampered := result
	tampered.exactCount++
	if tampered.Validate() == nil {
		t.Fatal("count tamper was accepted")
	}
	tampered = result
	tampered.stateSeal[0] ^= 0xff
	if tampered.Validate() == nil {
		t.Fatal("seal tamper was accepted")
	}
}
