package adoption

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTransformRemovesRawQueryAndPathInstances(t *testing.T) {
	page := legacyPage()
	records, targetSHA, err := TransformPage(manifest(LegacyTrace, 1), page)
	if err != nil || len(records) != 1 || targetSHA == [32]byte{} {
		t.Fatalf("transform failed: records=%+v err=%v", records, err)
	}
	if records[0].RoutePattern != "/api/forms/{id}" {
		t.Fatalf("route was not canonicalized: %q", records[0].RoutePattern)
	}
	document, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(document)
	for _, prohibited := range []string{"token=secret", "550e8400-e29b-41d4-a716-446655440000",
		"ip_address", "user_agent", "request_body", "response_body", "authorization"} {
		if strings.Contains(strings.ToLower(serialized), prohibited) {
			t.Fatalf("adoption payload contains prohibited raw data %q: %s", prohibited, serialized)
		}
	}
}

func TestTransformUsesStableV7TargetIdentity(t *testing.T) {
	page := legacyPage()
	first, firstSHA, err := TransformPage(manifest(LegacyTrace, 1), page)
	if err != nil {
		t.Fatal(err)
	}
	second, secondSHA, err := TransformPage(manifest(LegacyTrace, 1), page)
	if err != nil || first[0].TargetID != second[0].TargetID || firstSHA != secondSHA {
		t.Fatalf("transform is not deterministic: first=%+v second=%+v err=%v", first, second, err)
	}
	if first[0].TargetID[14] != '7' {
		t.Fatalf("target identity is not UUIDv7: %s", first[0].TargetID)
	}
}
