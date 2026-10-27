package aggregatepage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
)

type Page struct {
	groups    [][]byte
	limit     int
	stateSeal [sha256.Size]byte
	valid     bool
}

func newPage(groups [][]byte, limit int) (Page, error) {
	page := Page{groups: cloneGroups(groups), limit: limit, valid: true}
	page.stateSeal = pageSeal(page)
	if page.Validate() != nil {
		return Page{}, ErrInvalidPage
	}
	return page, nil
}

func (page Page) Validate() error {
	if !page.valid || page.groups == nil || page.limit < 1 || page.limit > 500 ||
		len(page.groups) > page.limit || page.stateSeal == [sha256.Size]byte{} ||
		pageSeal(page) != page.stateSeal {
		return ErrInvalidPage
	}
	for _, group := range page.groups {
		if !validGroup(group) {
			return ErrInvalidPage
		}
	}
	return nil
}

func (page Page) Groups() ([][]byte, error) {
	if page.Validate() != nil {
		return nil, ErrInvalidPage
	}
	return cloneGroups(page.groups), nil
}

func (page Page) Limit() (int, error) {
	if page.Validate() != nil {
		return 0, ErrInvalidPage
	}
	return page.limit, nil
}

func validGroup(document []byte) bool {
	if len(document) < 2 || document[0] != '{' || document[len(document)-1] != '}' {
		return false
	}
	scan, err := jsonscan.Scan(document, jsonscan.MaximumNesting)
	if err != nil || len(scan.DuplicatePointers()) != 0 {
		return false
	}
	var decoded map[string]json.RawMessage
	if json.Unmarshal(document, &decoded) != nil || len(decoded) != 2 {
		return false
	}
	return jsonObject(decoded["key"]) && jsonObject(decoded["measures"])
}

func jsonObject(document []byte) bool {
	var decoded map[string]json.RawMessage
	return json.Unmarshal(document, &decoded) == nil && decoded != nil
}

func (Page) String() string   { return redacted }
func (Page) GoString() string { return redacted }
func (Page) LogValue() slog.Value {
	return slog.StringValue(redacted)
}
func (page Page) Format(state fmt.State, _ rune) { _, _ = state.Write([]byte(redacted)) }
func (Page) MarshalJSON() ([]byte, error)        { return nil, ErrDirectMarshal }
