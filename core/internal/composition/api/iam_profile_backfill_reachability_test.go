package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIAMProfileBackfillPublicationIsReachableFromAPIComposition(t *testing.T) {
	runtime := readProfileBackfillSource(t, "iam_api_runtime.go")
	profileHTTP := readProfileBackfillSource(t, "iam_profile_http.go")
	publish := readProfileBackfillSource(t, filepath.Join("..", "..", "modules", "iam",
		"profile_http_schema_publish_application.go"))
	if !strings.Contains(runtime, "NewProfileBackfillPublicationRuntime(") ||
		!strings.Contains(profileHTTP, "runtime.profileBackfills") {
		t.Fatal("API composition does not inject profile publication backfill runtime")
	}
	replay := strings.Index(publish, "application.backfills.Replay(")
	originate := strings.Index(publish, "application.backfills.Originate(")
	direct := strings.Index(publish, "application.commands.Publish(")
	if replay < 0 || originate <= replay || direct <= originate {
		t.Fatalf("publish order replay/originate/direct = %d/%d/%d",
			replay, originate, direct)
	}
}

func readProfileBackfillSource(t *testing.T, path string) string {
	t.Helper()
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(document)
}
