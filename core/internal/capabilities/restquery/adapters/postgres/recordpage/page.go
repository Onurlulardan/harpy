package recordpage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
)

// Page is an immutable collection of canonical record JSON fragments.
type Page struct {
	records       [][]byte
	nextCursor    string
	requestedSize int
	hasMore       bool
	valid         bool
	stateSeal     [sha256.Size]byte
}

func newPage(records [][]byte, hasMore bool, nextCursor string, requested int) (Page, error) {
	page := Page{
		records: cloneRecords(records), nextCursor: nextCursor,
		requestedSize: requested, hasMore: hasMore, valid: true,
	}
	page.stateSeal = sealPage(page)
	if err := page.Validate(); err != nil {
		return Page{}, err
	}
	return page, nil
}

func (page Page) Validate() error {
	if !page.valid || page.requestedSize < 1 || page.requestedSize > 500 ||
		page.records == nil || len(page.records) > page.requestedSize ||
		page.hasMore != (page.nextCursor != "") ||
		page.hasMore && len(page.records) != page.requestedSize ||
		page.stateSeal == [sha256.Size]byte{} || sealPage(page) != page.stateSeal {
		return ErrInvalidPage
	}
	for _, record := range page.records {
		if !validJSONObject(record) {
			return ErrInvalidPage
		}
	}
	return nil
}

func (page Page) Records() ([][]byte, error) {
	if err := page.Validate(); err != nil {
		return nil, err
	}
	return cloneRecords(page.records), nil
}

func (page Page) HasMore() (bool, error) {
	if err := page.Validate(); err != nil {
		return false, err
	}
	return page.hasMore, nil
}

func (page Page) RequestedSize() (int, error) {
	if err := page.Validate(); err != nil {
		return 0, err
	}
	return page.requestedSize, nil
}

func (page Page) NextCursor() (string, bool, error) {
	if err := page.Validate(); err != nil {
		return "", false, err
	}
	return page.nextCursor, page.hasMore, nil
}

func (Page) String() string                      { return redacted }
func (Page) GoString() string                    { return redacted }
func (Page) LogValue() slog.Value                { return slog.StringValue(redacted) }
func (page Page) Format(state fmt.State, _ rune) { _, _ = state.Write([]byte(redacted)) }
func (Page) MarshalJSON() ([]byte, error)        { return nil, ErrDirectMarshal }
func validJSONObject(document []byte) bool {
	if len(document) < 2 || document[0] != '{' || document[len(document)-1] != '}' {
		return false
	}
	scan, err := jsonscan.Scan(document, jsonscan.MaximumNesting)
	if err != nil || len(scan.DuplicatePointers()) != 0 {
		return false
	}
	var value map[string]json.RawMessage
	return json.Unmarshal(document, &value) == nil && value != nil
}
