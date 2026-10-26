package analytics

import "testing"

func TestAnalyticsJobBindingRejectsUnknownKind(t *testing.T) {
	if _, err := analyticsJobBinding("analytics.unknown"); err == nil {
		t.Fatal("unknown Analytics job kind was accepted")
	}
}
