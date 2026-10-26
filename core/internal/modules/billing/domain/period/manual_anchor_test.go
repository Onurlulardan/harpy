package period

import (
	"errors"
	"testing"
	"time"
)

func TestMonthlyAnchorClampsWithoutDrifting(t *testing.T) {
	origin := time.Date(2024, time.January, 31, 9, 15, 30, 654321000, time.UTC)
	anchor, err := AnchorFrom(origin)
	if err != nil {
		t.Fatal(err)
	}
	first, err := BuildManual(origin, anchor, Monthly)
	if err != nil {
		t.Fatal(err)
	}
	wantFebruary := time.Date(2024, time.February, 29, 9, 15, 30, 654321000, time.UTC)
	if !first.End().Equal(wantFebruary) {
		t.Fatalf("January period end = %s", first.End())
	}
	second, err := BuildManual(first.End(), anchor, Monthly)
	if err != nil {
		t.Fatal(err)
	}
	wantMarch := time.Date(2024, time.March, 31, 9, 15, 30, 654321000, time.UTC)
	if !second.End().Equal(wantMarch) {
		t.Fatalf("February period end = %s", second.End())
	}
	if first.End().Sub(first.Start()) == 30*24*time.Hour ||
		second.End().Sub(second.Start()) == 30*24*time.Hour {
		t.Fatal("calendar month was reduced to 30 days")
	}
}

func TestYearlyLeapAnchorRecoversFebruary29(t *testing.T) {
	origin := time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)
	anchor, _ := AnchorFrom(origin)
	period, err := BuildManual(origin, anchor, Yearly)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2025, time.February, 28, 0, 0, 0, 0, time.UTC)
	if !period.End().Equal(want) {
		t.Fatalf("leap anchor end = %s", period.End())
	}
	for year := 2026; year <= 2028; year++ {
		period, err = BuildManual(period.End(), anchor, Yearly)
		if err != nil {
			t.Fatal(err)
		}
	}
	want = time.Date(2028, time.February, 29, 0, 0, 0, 0, time.UTC)
	if !period.End().Equal(want) {
		t.Fatalf("recovered leap anchor end = %s", period.End())
	}
}

func TestBuildManualRejectsAnchorMismatch(t *testing.T) {
	origin := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	anchor, _ := AnchorFrom(origin)
	wrong := time.Date(2026, time.January, 30, 0, 0, 0, 0, time.UTC)
	if _, err := BuildManual(wrong, anchor, Monthly); !errors.Is(err, ErrAnchor) {
		t.Fatalf("anchor mismatch error = %v", err)
	}
}
