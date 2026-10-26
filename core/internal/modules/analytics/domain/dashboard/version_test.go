package dashboard

import "testing"

func TestDashboardPublishRequiresExactNonEmptyImmutableWidgetSet(t *testing.T) {
	dashboardID, _ := NewID(v7(1))
	versionID, _ := NewVersionID(v7(2))
	widgetID, _ := NewID(v7(3))
	digest := [32]byte{1}
	widget, err := NewWidget(widgetID, v7(4), 0, digest, digest)
	if err != nil {
		t.Fatal(err)
	}
	version, err := NewDraft(versionID, dashboardID, 1, 7, VisibilityTenant, []Widget{widget})
	if err != nil {
		t.Fatal(err)
	}
	published, err := version.Publish(7)
	if err != nil || !published.Published() || published.Revision() != 8 {
		t.Fatalf("publish result = %#v, %v", published, err)
	}
	if _, err := published.Publish(8); err != ErrConflict {
		t.Fatalf("republish error = %v", err)
	}
}

func v7(seed byte) [16]byte {
	value := [16]byte{0, 0, 0, 0, 0, 0, 0x70, seed, 0x80}
	value[15] = seed
	return value
}
