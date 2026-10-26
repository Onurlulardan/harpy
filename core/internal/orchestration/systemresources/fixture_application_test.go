package systemresources

import "testing"

type testApplication struct {
	application *Application
	coordinator *testCoordinator
	iam         *testIAMPort
	layout      *testResolver
	fixture     testFixture
}

func newTestApplication(t *testing.T) testApplication {
	t.Helper()
	fixture := newTestFixture(t)
	coordinator := &testCoordinator{}
	iam := &testIAMPort{fixture: fixture}
	layout := &testResolver{document: fixture.document}
	application, err := NewApplication(coordinator, iam, layout)
	if err != nil {
		t.Fatal(err)
	}
	return testApplication{
		application: application, coordinator: coordinator,
		iam: iam, layout: layout, fixture: fixture,
	}
}

func (fixture testApplication) assertOneCommittedScope(t *testing.T) {
	t.Helper()
	if fixture.coordinator.calls != 1 || !fixture.coordinator.committed ||
		fixture.layout.calls != 1 || fixture.iam.scope != fixture.layout.scope {
		t.Fatalf(
			"scope calls=%d committed=%v layout=%d same=%v",
			fixture.coordinator.calls, fixture.coordinator.committed,
			fixture.layout.calls, fixture.iam.scope == fixture.layout.scope,
		)
	}
}
