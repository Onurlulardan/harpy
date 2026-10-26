package discount

import (
	"strings"
	"unicode/utf8"
)

const (
	maximumPresentationNameBytes        = 100
	maximumPresentationDescriptionBytes = 4096
)

type Presentation struct {
	name        string
	description string
}

func NewPresentation(name, description string) (Presentation, error) {
	if name == "" || strings.TrimSpace(name) != name || !utf8.ValidString(name) ||
		len(name) > maximumPresentationNameBytes || !utf8.ValidString(description) ||
		len(description) > maximumPresentationDescriptionBytes {
		return Presentation{}, ErrPolicy
	}
	return Presentation{name: name, description: description}, nil
}

func (value Presentation) Name() string        { return value.name }
func (value Presentation) Description() string { return value.description }

func (value Presentation) Valid() bool {
	verified, err := NewPresentation(value.name, value.description)
	return err == nil && verified == value
}
