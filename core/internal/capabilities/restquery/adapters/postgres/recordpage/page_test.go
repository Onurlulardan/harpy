package recordpage

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestPageIsImmutableRedactedAndRejectsDirectMarshal(t *testing.T) {
	source := [][]byte{[]byte(`{"name":"Ada"}`)}
	page, err := newPage(source, false, "", 2)
	if err != nil {
		t.Fatal(err)
	}
	source[0][2] = 'X'
	records, err := page.Records()
	if err != nil || string(records[0]) != `{"name":"Ada"}` {
		t.Fatalf("records=%q err=%v", records, err)
	}
	records[0][2] = 'Y'
	again, err := page.Records()
	if err != nil || string(again[0]) != `{"name":"Ada"}` {
		t.Fatal("record accessor exposed mutable state")
	}
	if fmt.Sprint(page) != redacted || page.LogValue().String() != redacted {
		t.Fatal("page formatting leaked state")
	}
	if _, err := json.Marshal(page); !errors.Is(err, ErrDirectMarshal) {
		t.Fatalf("direct marshal error=%v", err)
	}
}

func TestPageDetectsTampering(t *testing.T) {
	base, err := newPage([][]byte{[]byte(`{"id":"one"}`)}, false, "", 1)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func(*Page){
		"record": func(page *Page) { page.records[0][2] ^= 1 },
		"cursor": func(page *Page) { page.nextCursor = "forged" },
		"more":   func(page *Page) { page.hasMore = true },
		"size":   func(page *Page) { page.requestedSize++ },
		"seal":   func(page *Page) { page.stateSeal[0] ^= 1 },
		"valid":  func(page *Page) { page.valid = false },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			page := base
			page.records = cloneRecords(base.records)
			mutate(&page)
			if page.Validate() != ErrInvalidPage {
				t.Fatal("tampered page was accepted")
			}
		})
	}
}

func TestPageAllowsEmptyAndRejectsIncoherentPagination(t *testing.T) {
	page, err := newPage([][]byte{}, false, "", 25)
	if err != nil {
		t.Fatal(err)
	}
	more, moreErr := page.HasMore()
	requested, requestedErr := page.RequestedSize()
	cursor, present, cursorErr := page.NextCursor()
	if moreErr != nil || requestedErr != nil || cursorErr != nil || requested != 25 ||
		more || present || cursor != "" {
		t.Fatal("empty page exposed continuation")
	}
	if _, err := newPage([][]byte{}, true, "token", 25); err != ErrInvalidPage {
		t.Fatalf("incoherent page error=%v", err)
	}
}

func TestPageRejectsDuplicateJSONMembers(t *testing.T) {
	for name, document := range map[string][]byte{
		"root":   []byte(`{"name":"Ada","name":"Grace"}`),
		"nested": []byte(`{"data":{"id":"one","id":"two"}}`),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := newPage([][]byte{document}, false, "", 1); err != ErrInvalidPage {
				t.Fatalf("duplicate member error=%v", err)
			}
		})
	}
}

func TestPageSealHasRecordPageDomainSeparation(t *testing.T) {
	page, err := newPage([][]byte{[]byte(`{"id":"one"}`)}, false, "", 1)
	if err != nil {
		t.Fatal(err)
	}
	legacy := legacyUndomainedPageSeal(page)
	if pageSealDomain != "forgecrud.restquery.record-page.v1" || page.stateSeal == legacy {
		t.Fatal("page seal is not domain separated")
	}
}

func legacyUndomainedPageSeal(page Page) [sha256.Size]byte {
	hash := sha256.New()
	writeUint64(hash.Write, uint64(page.requestedSize))
	_, _ = hash.Write([]byte{0})
	writeBytes(hash.Write, []byte(page.nextCursor))
	writeUint64(hash.Write, uint64(len(page.records)))
	for _, record := range page.records {
		writeBytes(hash.Write, record)
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}
