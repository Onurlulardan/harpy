package main

import "testing"

func TestParseRelease(t *testing.T) {
	release, err := parseRelease("1.27.3")
	if err != nil {
		t.Fatal(err)
	}
	if release.String() != "1.27.3" || release.Toolchain() != "go1.27.3" ||
		release.Language() != "1.27.0" || release.Minor() != "1.27" {
		t.Fatalf("parsed release = %#v", release)
	}
	for _, raw := range []string{"", "1.27", "latest", "v1.27.3", "1.27.03", "0.27.3"} {
		if _, err := parseRelease(raw); err == nil {
			t.Fatalf("accepted invalid version %q", raw)
		}
	}
}
