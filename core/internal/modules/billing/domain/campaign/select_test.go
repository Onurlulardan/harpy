package campaign

import (
	"errors"
	"testing"
	"time"
)

func TestSelectActiveUsesPriorityThenIdentifier(t *testing.T) {
	at := instant(12)
	lower := version(t, identifier(9), 3, at.Add(-time.Hour), nil, nil)
	tieLater := version(t, identifier(8), 7, at.Add(-time.Hour), nil, nil)
	tieEarlier := version(t, identifier(7), 7, at.Add(-time.Hour), nil, nil)

	selection, err := SelectActive(at, []Version{lower, tieLater, tieEarlier})
	if err != nil {
		t.Fatal(err)
	}
	if selection == nil || selection.Version().Identifier() != identifier(7) {
		t.Fatalf("unexpected selection: %#v", selection)
	}
}

func TestSelectActiveUsesHalfOpenBoundaryAndFiltersDiscounts(t *testing.T) {
	at := instant(12)
	end := at
	expired := discount(t, identifier(3), 1, true, at.Add(-time.Hour), &end)
	future := discount(t, identifier(4), 1, true, at.Add(time.Hour), nil)
	revoked := discount(t, identifier(5), 1, false, at.Add(-time.Hour), nil)
	tieLater := discount(t, identifier(7), 2, true, at.Add(-time.Hour), nil)
	tieEarlier := discount(t, identifier(6), 2, true, at.Add(-time.Hour), nil)
	first := discount(t, identifier(8), 1, true, at.Add(-time.Hour), nil)
	selected := version(t, identifier(1), 1, at.Add(-time.Hour), nil,
		[]Discount{expired, future, revoked, tieLater, tieEarlier, first})
	ending := version(t, identifier(2), 99, at.Add(-time.Hour), &end, nil)

	selection, err := SelectActive(at, []Version{ending, selected})
	if err != nil {
		t.Fatal(err)
	}
	discounts := selection.Discounts()
	want := []Identifier{identifier(8), identifier(6), identifier(7)}
	if len(discounts) != len(want) {
		t.Fatalf("got %d discounts", len(discounts))
	}
	for index := range want {
		if discounts[index].Identifier() != want[index] {
			t.Fatalf("discount %d mismatch", index)
		}
	}
}

func TestSelectActiveReturnsNilAndRejectsNonCanonicalInstant(t *testing.T) {
	selection, err := SelectActive(instant(12), nil)
	if err != nil || selection != nil {
		t.Fatalf("selection=%#v err=%v", selection, err)
	}
	_, err = SelectActive(instant(12).Add(time.Nanosecond), nil)
	if !errors.Is(err, ErrInstant) {
		t.Fatalf("expected ErrInstant, got %v", err)
	}
}

func identifier(last byte) Identifier {
	var value Identifier
	value[15] = last
	return value
}

func instant(hour int) time.Time {
	return time.Date(2026, time.July, 17, hour, 0, 0, 0, time.UTC)
}

func discount(
	t *testing.T, id Identifier, priority int32, active bool,
	start time.Time, end *time.Time,
) Discount {
	t.Helper()
	value, err := NewDiscount(id, priority, active, start, end)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func version(
	t *testing.T, id Identifier, priority int32, start time.Time,
	end *time.Time, discounts []Discount,
) Version {
	t.Helper()
	value, err := NewVersion(id, identifier(99), priority, true, true,
		start, end, discounts)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
