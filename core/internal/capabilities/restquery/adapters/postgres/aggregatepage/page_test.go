package aggregatepage

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestPageIsImmutableRedactedAndRejectsDirectMarshal(t *testing.T) {
	source := [][]byte{[]byte(`{"key":{"status":"OPEN"},"measures":{"count":"1"}}`)}
	page, err := newPage(source, 2)
	if err != nil {
		t.Fatal(err)
	}
	source[0][2] = 'X'
	groups, err := page.Groups()
	if err != nil || groups[0][2] != 'k' {
		t.Fatal("page retained caller-owned input")
	}
	groups[0][2] = 'Y'
	again, _ := page.Groups()
	if again[0][2] != 'k' {
		t.Fatal("group accessor exposed mutable state")
	}
	if fmt.Sprint(page) != redacted || fmt.Sprintf("%#v", page) != redacted ||
		page.LogValue().String() != redacted {
		t.Fatal("page formatting leaked state")
	}
	if _, err := json.Marshal(page); !errors.Is(err, ErrDirectMarshal) {
		t.Fatalf("direct marshal error = %v", err)
	}
}

func TestPageRejectsDuplicateMissingAndExtraMembers(t *testing.T) {
	for name, document := range map[string][]byte{
		"duplicate root": []byte(
			`{"key":{},"key":{},"measures":{}}`,
		),
		"duplicate nested": []byte(
			`{"key":{"status":"A","status":"B"},"measures":{}}`,
		),
		"missing key":         []byte(`{"measures":{}}`),
		"missing measures":    []byte(`{"key":{}}`),
		"extra root":          []byte(`{"key":{},"measures":{},"extra":true}`),
		"non-object key":      []byte(`{"key":[],"measures":{}}`),
		"non-object measures": []byte(`{"key":{},"measures":null}`),
		"trailing bytes":      []byte(`{"key":{},"measures":{}}x`),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := newPage([][]byte{document}, 1); err != ErrInvalidPage {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestPageDetectsTamperingAndRejectsZeroValue(t *testing.T) {
	base, err := newPage(
		[][]byte{[]byte(`{"key":{},"measures":{"count":"1"}}`)}, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*Page){
		"group": func(page *Page) { page.groups[0][2] ^= 1 },
		"limit": func(page *Page) { page.limit++ },
		"seal":  func(page *Page) { page.stateSeal[0] ^= 1 },
		"valid": func(page *Page) { page.valid = false },
	} {
		t.Run(name, func(t *testing.T) {
			page := base
			page.groups = cloneGroups(base.groups)
			mutate(&page)
			if page.Validate() != ErrInvalidPage {
				t.Fatal("tampered page accepted")
			}
		})
	}
	if (Page{}).Validate() != ErrInvalidPage {
		t.Fatal("zero page accepted")
	}
}

func TestMeasureJSONAllowsOnlyCanonicalStringOrNull(t *testing.T) {
	for _, accepted := range [][]byte{[]byte(`"0"`), []byte(`"-1.25"`), []byte("null")} {
		if !measureJSON(accepted) {
			t.Fatalf("canonical measure rejected: %s", accepted)
		}
	}
	for _, rejected := range [][]byte{nil, []byte("0"), []byte("true"), []byte(`{}`)} {
		if measureJSON(rejected) {
			t.Fatalf("non-string measure accepted: %s", rejected)
		}
	}
}
