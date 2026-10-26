package resolved

import (
	"fmt"
	"unicode/utf8"
)

const maximumResolvedCursorBytes = 16 * 1024

type Page struct {
	size      int
	after     string
	hasAfter  bool
	validated bool
}

func NewPage(size int, after *string) (Page, error) {
	page := Page{size: size, validated: true}
	if after != nil {
		page.after, page.hasAfter = *after, true
	}
	if err := page.Validate(); err != nil {
		return Page{}, err
	}
	return page, nil
}

func (page Page) Validate() error {
	if !page.validated || page.size < 1 || page.size > 500 ||
		page.hasAfter != (page.after != "") ||
		page.hasAfter && (!utf8.ValidString(page.after) || len(page.after) > maximumResolvedCursorBytes) {
		return fmt.Errorf("resolved REST Query page is invalid")
	}
	return nil
}

func (page Page) Size() (int, error) {
	if err := page.Validate(); err != nil {
		return 0, err
	}
	return page.size, nil
}

func (page Page) After() (string, bool, error) {
	if err := page.Validate(); err != nil {
		return "", false, err
	}
	return page.after, page.hasAfter, nil
}

func (Page) String() string                      { return valueRedaction }
func (Page) GoString() string                    { return valueRedaction }
func (page Page) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Page) MarshalJSON() ([]byte, error)        { return rejectDirectJSONMarshal() }
