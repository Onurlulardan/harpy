package v1query

import (
	"context"
	"slices"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

func TestCoordinatorOrdersAuthorizationEvidenceAndMinting(t *testing.T) {
	events := []string{}
	evidence := &coordinatorEvidence{events: &events}
	coordinator, err := NewCoordinator(
		staticManifestSource{snapshot: testManifest()},
		coordinatorAuthorizer{t: t, events: &events}, evidence,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); err != nil {
		t.Fatal(err)
	}
	events = nil
	accepted, err := coordinator.Accept(
		context.Background(),
		"q=term&filters[legacy_status][operator]=same_as&"+
			"filters[legacy_status][value]=OPEN&sort[field]=legacy_status",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(events, []string{"authorize", "hit"}) {
		t.Fatalf("events=%v", events)
	}
	if accepted.Validate() != nil || evidence.hit != 1 {
		t.Fatal("accepted execution or evidence is invalid")
	}
	if _, _, err := accepted.ConsumeExecution(); err != nil {
		t.Fatal(err)
	}
	warning, present, err := accepted.Warning()
	if err != nil || !present || warning != DeprecatedWarningValue ||
		WarningHeaderName != "Warning" {
		t.Fatalf("warning=(%q,%v,%v)", warning, present, err)
	}
	if _, _, err := accepted.ConsumeExecution(); err == nil {
		t.Fatal("accepted execution was reusable")
	}
}

func TestCoordinatorMapsRetiredAliasesToOriginalUnknownIssues(t *testing.T) {
	events := []string{}
	evidence := &coordinatorEvidence{events: &events}
	coordinator, err := NewCoordinator(
		staticManifestSource{snapshot: testManifest()},
		coordinatorAuthorizer{t: t, events: &events}, evidence,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); err != nil {
		t.Fatal(err)
	}
	evidence.retired = []contractretirement.Key{
		coordinatorKey(t, "query.search.v1"),
		coordinatorKey(t, "field.status.v1"),
		coordinatorKey(t, "operator.same-as.v1"),
		coordinatorKey(t, "sort.status.v1"),
	}
	events = nil
	_, err = coordinator.Accept(
		context.Background(),
		"q=term&filters[legacy_status][operator]=same_as&"+
			"filters[legacy_status][value]=OPEN&sort[field]=legacy_status",
	)
	requireQueryError(
		t, err, restqueryv1.ErrorQueryInvalid, "/q",
		restqueryv1.IssueMemberUnknown,
	)
	requireQueryError(
		t, err, restqueryv1.ErrorQueryInvalid, "/filters/legacy_status",
		restqueryv1.IssueFieldUnknown,
	)
	requireQueryError(
		t, err, restqueryv1.ErrorQueryInvalid,
		"/filters/legacy_status/operator", restqueryv1.IssueOperatorUnknown,
	)
	requireQueryError(
		t, err, restqueryv1.ErrorQueryInvalid, "/sort/field",
		restqueryv1.IssueFieldUnknown,
	)
	if !slices.Equal(events, []string{"authorize", "hit"}) {
		t.Fatalf("events=%v", events)
	}
}

func TestCoordinatorSkipsEvidenceWithoutAliasAndAdmitsFullManifest(t *testing.T) {
	events := []string{}
	evidence := &coordinatorEvidence{events: &events}
	coordinator, err := NewCoordinator(
		staticManifestSource{snapshot: testManifest()},
		coordinatorAuthorizer{t: t, events: &events}, evidence,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); err != nil {
		t.Fatal(err)
	}
	accepted, err := coordinator.Accept(context.Background(), "search=term")
	if err != nil || evidence.hit != 0 {
		t.Fatalf("accept err=%v hits=%d", err, evidence.hit)
	}
	if warning, present, _ := accepted.Warning(); present || warning != "" {
		t.Fatal("canonical request emitted deprecation warning")
	}
	if evidence.admit != 1 || evidence.hit != 0 {
		t.Fatalf("admit=%d hit=%d", evidence.admit, evidence.hit)
	}
}

func TestCoordinatorRejectsRetiredManifestAndInvalidPorts(t *testing.T) {
	events := []string{}
	evidence := &coordinatorEvidence{
		events:  &events,
		retired: []contractretirement.Key{coordinatorKey(t, "query.search.v1")},
	}
	coordinator, err := NewCoordinator(
		staticManifestSource{snapshot: testManifest()},
		coordinatorAuthorizer{t: t, events: &events}, evidence,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); !contractretirement.FaultIs(err, contractretirement.FaultRetired) {
		t.Fatalf("admit err=%v", err)
	}
	readyEvidence := &coordinatorEvidence{events: &events}
	notAdmitted, err := NewCoordinator(
		staticManifestSource{snapshot: testManifest()},
		coordinatorAuthorizer{t: t, events: &events}, readyEvidence,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := notAdmitted.Accept(context.Background(), "search=term"); err == nil {
		t.Fatal("request was accepted before manifest admission")
	}
	var nilEvidence *coordinatorEvidence
	if _, err := NewCoordinator(
		staticManifestSource{snapshot: testManifest()},
		coordinatorAuthorizer{t: t, events: &events}, nilEvidence,
	); err == nil {
		t.Fatal("typed nil evidence was accepted")
	}
	if (Accepted{}).Validate() == nil {
		t.Fatal("zero accepted execution validated")
	}
}
