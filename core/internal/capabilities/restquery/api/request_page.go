package api

import (
	"fmt"
	"unicode/utf8"
)

const (
	maximumPageSize    = 500
	maximumCursorBytes = 16 * 1024
)

type RequestPage struct {
	size     int
	after    string
	hasAfter bool
}

func NewRequestPage(size int, after *string) (RequestPage, error) {
	page := RequestPage{size: size}
	if after != nil {
		page.after = *after
		page.hasAfter = true
	}
	if !page.valid() {
		return RequestPage{}, fmt.Errorf("REST Query page is invalid")
	}
	return page, nil
}

func (page RequestPage) Size() int { return page.size }

func (page RequestPage) After() (string, bool) {
	return page.after, page.hasAfter
}

func (page RequestPage) String() string   { return queryValueRedaction }
func (page RequestPage) GoString() string { return queryValueRedaction }
func (page RequestPage) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestPage) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (page RequestPage) valid() bool {
	return page.size >= 1 && page.size <= maximumPageSize &&
		(!page.hasAfter || page.after != "" && utf8.ValidString(page.after) &&
			len(page.after) <= maximumCursorBytes)
}
