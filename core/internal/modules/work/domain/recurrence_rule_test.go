package domain

import (
	"testing"
	"time"
)

func TestParseRecurrenceRejectsUnknownAndUnboundedGrammar(t *testing.T) {
	location, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		t.Fatal(err)
	}
	base := RecurrenceInput{DTStart: "20260301T090000", TimeZone: "Europe/Istanbul",
		TZDBVersion: "2026a", RRULE: "FREQ=WEEKLY;BYDAY=MO,WE", Location: location}
	if _, err := ParseRecurrence(base); err != nil {
		t.Fatal(err)
	}
	for _, rule := range []string{
		"FREQ=HOURLY", "FREQ=DAILY;BYSECOND=1", "FREQ=DAILY;COUNT=2;UNTIL=20260303T090000",
	} {
		base.RRULE = rule
		if _, err := ParseRecurrence(base); err == nil {
			t.Fatalf("accepted %q", rule)
		}
	}
}

func TestMaterializePinsTZDBAndSkipsNonexistentDSTTime(t *testing.T) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := ParseRecurrence(RecurrenceInput{DTStart: "20260307T023000",
		TimeZone: "America/New_York", TZDBVersion: "2026a", RRULE: "FREQ=DAILY;COUNT=3",
		Location: location})
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)
	occurrences, err := Materialize(definition, from, from.Add(5*24*time.Hour), 10)
	if err != nil || len(occurrences) != 3 {
		t.Fatalf("occurrences = %+v, %v", occurrences, err)
	}
	for _, occurrence := range occurrences {
		if occurrence.TZDBVersion != "2026a" || occurrence.LocalLabel == "2026-03-08T02:30:00" {
			t.Fatalf("unpinned or nonexistent occurrence: %+v", occurrence)
		}
	}
}
