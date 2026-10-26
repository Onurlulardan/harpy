package v1query

import (
	"context"
	"slices"
	"testing"
)

func TestPreparedAdmitsThenAuthorizesThenRecordsAliasHits(t *testing.T) {
	events := []string{}
	evidence := &coordinatorEvidence{events: &events}
	prepared, err := Prepare(
		context.Background(), staticManifestSource{snapshot: testManifest()}, evidence,
		DialectDynamicList,
		"q=term&filters[legacy_status][operator]=same_as&"+
			"filters[legacy_status][value]=OPEN&sort[field]=legacy_status",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(events, []string{"admit"}) {
		t.Fatalf("prepare events=%v", events)
	}
	accepted, err := prepared.Accept(
		context.Background(), coordinatorAuthorizer{t: t, events: &events},
	)
	if err != nil || accepted.Validate() != nil {
		t.Fatalf("accept error=%v", err)
	}
	if !slices.Equal(events, []string{"admit", "authorize", "hit"}) {
		t.Fatalf("events=%v", events)
	}
	warning, deprecated, err := accepted.Warning()
	if err != nil || !deprecated || warning != DeprecatedWarningValue {
		t.Fatal("accepted aliases did not carry the versioned warning")
	}
	if _, err = prepared.Accept(
		context.Background(), coordinatorAuthorizer{t: t, events: &events},
	); err == nil {
		t.Fatal("prepared query was accepted more than once")
	}
}

func TestPreparedExposesPagePlanWithoutConsumingAuthorization(t *testing.T) {
	events := []string{}
	prepared, err := Prepare(
		context.Background(), staticManifestSource{snapshot: testManifest()},
		&coordinatorEvidence{events: &events}, DialectOpenAPIGet,
		"page=4&limit=25&filter[status]=OPEN",
	)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := prepared.PagePlan()
	if err != nil {
		t.Fatal(err)
	}
	page, _ := plan.Page()
	limit, _ := plan.Limit()
	predecessor, present, _ := plan.PredecessorIndex()
	if page != 4 || limit != 25 || !present || predecessor != 74 {
		t.Fatalf("plan=(%d,%d,%d,%v)", page, limit, predecessor, present)
	}
}
