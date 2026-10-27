package campaign

import (
	"strings"
	"testing"
	"time"
)

func TestPublicationCanonicalizesIdentifiersAndDefensivelyCopies(t *testing.T) {
	start := instant(12)
	end := start.Add(time.Hour)
	identifiers := []Identifier{identifier(3), identifier(1), identifier(2)}
	publication, err := NewPublication(
		"Yaz Kampanyası", "Sadakat kampanyası", "Üyelere özel", true,
		-7, start, &end, identifiers,
	)
	if err != nil || !publication.Valid() {
		t.Fatalf("publication=%#v err=%v", publication, err)
	}
	identifiers[0] = Identifier{}
	end = start
	got := publication.DiscountVersionIDs()
	if got[0] != identifier(1) || got[2] != identifier(3) ||
		publication.Description() != "Sadakat kampanyası" || !publication.AutoApply() ||
		publication.EndsAt() == nil || !publication.EndsAt().After(start) {
		t.Fatalf("publication changed through caller input: %#v", publication)
	}
	got[0] = Identifier{}
	returnedEnd := publication.EndsAt()
	*returnedEnd = start
	if publication.DiscountVersionIDs()[0] != identifier(1) ||
		!publication.EndsAt().After(start) {
		t.Fatal("publication changed through projection")
	}
}

func TestPublicationAcceptsEmptyDiscountSet(t *testing.T) {
	publication, err := NewPublication("Duyuru", "", "", false, 0, instant(12), nil, nil)
	if err != nil || len(publication.DiscountVersionIDs()) != 0 {
		t.Fatalf("publication=%#v err=%v", publication, err)
	}
}

func TestPublicationRejectsPresentationAndIdentifierViolations(t *testing.T) {
	start := instant(12)
	invalidUTF8 := string([]byte{0xff})
	cases := []struct {
		name        string
		description string
		banner      string
		discounts   []Identifier
	}{
		{"", "", "", nil},
		{" padded ", "", "", nil},
		{strings.Repeat("x", 129), "", "", nil},
		{invalidUTF8, "", "", nil},
		{"Valid", strings.Repeat("x", 4097), "", nil},
		{"Valid", invalidUTF8, "", nil},
		{"Valid", "", strings.Repeat("x", 513), nil},
		{"Valid", "", invalidUTF8, nil},
		{"Valid", "", "", []Identifier{{}}},
		{"Valid", "", "", []Identifier{identifier(1), identifier(1)}},
	}
	for _, test := range cases {
		if _, err := NewPublication(
			test.name, test.description, test.banner, false, 0, start, nil,
			test.discounts,
		); err != ErrPublication {
			t.Fatalf("case %#v: expected ErrPublication, got %v", test, err)
		}
	}
	tooMany := make([]Identifier, 257)
	for index := range tooMany {
		tooMany[index][14] = byte(index/255 + 1)
		tooMany[index][15] = byte(index%255 + 1)
	}
	if _, err := NewPublication("Valid", "", "", false, 0, start, nil, tooMany); err != ErrPublication {
		t.Fatalf("expected ErrPublication for 257 discounts, got %v", err)
	}
}

func TestPublicationRequiresUTCMicrosecondHalfOpenInterval(t *testing.T) {
	start := instant(12)
	end := start
	if _, err := NewPublication("Valid", "", "", false, 0, start, &end, nil); err != ErrPublication {
		t.Fatalf("empty interval accepted: %v", err)
	}
	if _, err := NewPublication(
		"Valid", "", "", false, 0, start.Add(time.Nanosecond), nil, nil,
	); err != ErrPublication {
		t.Fatalf("nanosecond instant accepted: %v", err)
	}
	offset := start.In(time.FixedZone("offset", 3*60*60))
	if _, err := NewPublication("Valid", "", "", false, 0, offset, nil, nil); err != ErrPublication {
		t.Fatalf("non-UTC instant accepted: %v", err)
	}
}
