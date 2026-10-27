package jsonscan

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"unicode/utf8"
)

var (
	ErrInvalidDocument = errors.New("JSON scan failed")
	ErrNestingLimit    = errors.New("JSON scan exceeded nesting limit")
)

const MaximumNesting = 128

type Result struct {
	duplicates []string
}

func (result Result) DuplicatePointers() []string {
	return append([]string(nil), result.duplicates...)
}

type scanState struct {
	maximumDepth int
	duplicates   map[string]struct{}
}

func Scan(document []byte, maximumDepth int) (Result, error) {
	if len(document) == 0 || !utf8.Valid(document) || maximumDepth < 1 {
		return Result{}, ErrInvalidDocument
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	state := scanState{
		maximumDepth: maximumDepth, duplicates: make(map[string]struct{}),
	}
	if err := scanValue(decoder, "", 0, &state); err != nil {
		return Result{}, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return Result{}, ErrInvalidDocument
	}
	return state.result(), nil
}

func (state scanState) result() Result {
	paths := make([]string, 0, len(state.duplicates))
	for path := range state.duplicates {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return Result{duplicates: paths}
}
