package published

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const maximumPublishedCursorBytes = 16 * 1024

type PageOverride struct {
	size      int
	after     string
	hasSize   bool
	hasAfter  bool
	validated bool
}

func NewPageOverride(size *int, after *string) (PageOverride, error) {
	page := PageOverride{validated: true}
	if size != nil {
		page.size, page.hasSize = *size, true
	}
	if after != nil {
		page.after, page.hasAfter = *after, true
	}
	if err := page.Validate(); err != nil {
		return PageOverride{}, err
	}
	return page, nil
}

func (page PageOverride) Size() (int, bool, error) {
	if err := page.Validate(); err != nil {
		return 0, false, err
	}
	return page.size, page.hasSize, nil
}

func (page PageOverride) After() (string, bool, error) {
	if err := page.Validate(); err != nil {
		return "", false, err
	}
	return page.after, page.hasAfter, nil
}

func (page PageOverride) Validate() error {
	if !page.validated || !page.hasSize && !page.hasAfter ||
		page.hasSize && (page.size < 1 || page.size > 500) ||
		page.hasAfter && (page.after == "" || len(page.after) > maximumPublishedCursorBytes ||
			!utf8.ValidString(page.after) || strings.ContainsRune(page.after, '\x00')) {
		return fmt.Errorf("published REST Query page override is invalid")
	}
	return nil
}

func (PageOverride) String() string   { return publishedRedaction }
func (PageOverride) GoString() string { return publishedRedaction }
func (page PageOverride) Format(state fmt.State, _ rune) {
	writePublishedRedaction(state)
}
func (PageOverride) MarshalJSON() ([]byte, error) { return rejectPublishedJSON() }
