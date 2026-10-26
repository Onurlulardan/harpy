package resolver

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestResolvedQuerySpecRedactsAndRejectsDirectJSON(t *testing.T) {
	fixture := newQueryFixture(t, true)
	spec, err := Resolve(fixtureRecordRequest(t, nil), fixture.set)
	if err != nil {
		t.Fatal(err)
	}
	for _, rendered := range []string{
		fmt.Sprint(spec), fmt.Sprintf("%v", spec), fmt.Sprintf("%+v", spec),
		fmt.Sprintf("%#v", spec),
	} {
		if rendered != "[REDACTED]" || strings.Contains(rendered, "amount") ||
			strings.Contains(rendered, "Ada") {
			t.Fatalf("resolved spec leaked through formatting: %q", rendered)
		}
	}
	if _, err := json.Marshal(spec); err == nil {
		t.Fatal("resolved spec allowed direct JSON marshaling")
	}
}

func TestResolvedQueryAccessorsAreDefensiveAndZeroFailsClosed(t *testing.T) {
	fixture := newQueryFixture(t, true)
	spec, err := Resolve(fixtureRecordRequest(t, nil), fixture.set)
	if err != nil {
		t.Fatal(err)
	}
	record, _, _ := spec.Record()
	projections, _ := record.Projections()
	projections[0] = resolved.Projection{}
	again, err := record.Projections()
	if err != nil || again[0].Validate() != nil {
		t.Fatal("projection accessor exposed mutable internal state")
	}
	includes, _ := record.Includes()
	includes[0] = resolved.Include{}
	againIncludes, err := record.Includes()
	if err != nil || againIncludes[0].Validate() != nil {
		t.Fatal("include accessor exposed mutable internal state")
	}
	zero := resolved.QuerySpec{}
	if zero.Validate() == nil {
		t.Fatal("zero QuerySpec validated")
	}
	if _, err := zero.Mode(); err == nil {
		t.Fatal("zero QuerySpec accessor did not fail closed")
	}
	if _, _, err := zero.Record(); err == nil {
		t.Fatal("zero QuerySpec exposed record state")
	}
}
