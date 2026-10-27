package recordpage

import (
	"bytes"
	"testing"
)

func TestAppendIncludedIsFinalAndRejectsReservedCollision(t *testing.T) {
	result, err := AppendIncluded(
		[]byte(`{"name":"Ada"}`),
		[]byte(`{"children":[]}`),
	)
	if err != nil || string(result) !=
		`{"name":"Ada","_included":{"children":[]}}` {
		t.Fatalf("result=%s error=%v", result, err)
	}
	if !bytes.HasSuffix(result, []byte(`"_included":{"children":[]}}`)) {
		t.Fatal("reserved member is not final")
	}
	if _, err := AppendIncluded(
		[]byte(`{"_included":{}}`), []byte(`{}`),
	); err == nil {
		t.Fatal("reserved member collision was accepted")
	}
}

func TestAttachIncludedPreservesPageMetadata(t *testing.T) {
	page, err := newPage([][]byte{[]byte(`{"name":"Ada"}`)}, false, "", 7)
	if err != nil {
		t.Fatal(err)
	}
	attached, err := AttachIncluded(page, [][]byte{[]byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	records, recordsErr := attached.Records()
	size, sizeErr := attached.RequestedSize()
	if recordsErr != nil || sizeErr != nil || size != 7 ||
		string(records[0]) != `{"name":"Ada","_included":{}}` {
		t.Fatalf("records=%q size=%d errors=%v/%v", records, size, recordsErr, sizeErr)
	}
}
