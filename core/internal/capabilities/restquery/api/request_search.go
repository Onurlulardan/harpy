package api

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	maximumSearchRunes = 128
	maximumSearchBytes = 256
)

type RequestSearch struct {
	term string
}

func NewRequestSearch(term string) (RequestSearch, error) {
	if !utf8.ValidString(term) {
		return RequestSearch{}, fmt.Errorf("REST Query search term is invalid UTF-8")
	}
	normalized := norm.NFC.String(term)
	search := RequestSearch{term: normalized}
	if !search.valid() {
		return RequestSearch{}, fmt.Errorf("REST Query search term is invalid")
	}
	return search, nil
}

func (search RequestSearch) Term() string          { return search.term }
func (search RequestSearch) String() string        { return queryValueRedaction }
func (search RequestSearch) GoString() string      { return queryValueRedaction }
func (RequestSearch) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
func (search RequestSearch) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}

func (search RequestSearch) valid() bool {
	return utf8.ValidString(search.term) && norm.NFC.IsNormalString(search.term) &&
		!strings.ContainsRune(search.term, '\x00') && strings.TrimSpace(search.term) != "" &&
		utf8.RuneCountInString(search.term) <= maximumSearchRunes &&
		len(search.term) <= maximumSearchBytes
}
