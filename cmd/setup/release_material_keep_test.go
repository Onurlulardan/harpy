package main

import (
	"testing"
	"time"
)

func TestReleaseCatalogMatchesBuild(t *testing.T) {
	files, err := buildReleaseFiles(newReleaseClock(time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	catalog := releaseCatalog()
	if len(files) != len(catalog) {
		t.Fatalf("release files = %d, catalog = %d", len(files), len(catalog))
	}
	for index := range files {
		if files[index].key != catalog[index].key || files[index].name != catalog[index].name ||
			len(files[index].body) == 0 {
			t.Fatalf("release catalog drifted at %s", catalog[index].name)
		}
	}
}
