package jsonscan

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestScanFindsSortedEscapedDuplicatesAtEveryDepth(t *testing.T) {
	result, err := Scan([]byte(
		`{"z":1,"z":2,"a/b~c":{"x":1,"x":2},"items":[{"q":1,"q":2}]}`,
	), 128)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/a~1b~0c/x", "/items/0/q", "/z"}
	if got := result.DuplicatePointers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("duplicates = %#v", got)
	}
	first := result.DuplicatePointers()
	first[0] = "changed"
	if reflect.DeepEqual(first, result.DuplicatePointers()) {
		t.Fatal("duplicate accessor exposed mutable state")
	}
}

func TestScanRejectsMalformedTrailingAndDepthOverflow(t *testing.T) {
	for _, document := range [][]byte{nil, []byte(`{`), []byte(`{} {}`), {0xff}} {
		if _, err := Scan(document, 128); !errors.Is(err, ErrInvalidDocument) {
			t.Fatalf("error = %v", err)
		}
	}
	if _, err := Scan([]byte(strings.Repeat("[", 129)+strings.Repeat("]", 129)), 128); !errors.Is(err, ErrNestingLimit) {
		t.Fatalf("depth error = %v", err)
	}
}
