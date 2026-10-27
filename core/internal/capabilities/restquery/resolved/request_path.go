package resolved

import (
	"fmt"
	"unicode/utf8"
)

type RequestPath struct {
	value     string
	validated bool
}

func NewRequestPath(value string) (RequestPath, error) {
	path := RequestPath{value: value, validated: true}
	if err := path.Validate(); err != nil {
		return RequestPath{}, err
	}
	return path, nil
}

func (path RequestPath) Validate() error {
	if !path.validated || !validRequestPath(path.value) {
		return fmt.Errorf("resolved REST Query request path is invalid")
	}
	return nil
}

func (path RequestPath) Value() (string, error) {
	if err := path.Validate(); err != nil {
		return "", err
	}
	return path.value, nil
}

func validRequestPath(value string) bool {
	if !utf8.ValidString(value) || len(value) > 256 {
		return false
	}
	if value == "" {
		return true
	}
	if value[0] != '/' {
		return false
	}
	for index := 1; index < len(value); index++ {
		if value[index] != '~' {
			continue
		}
		index++
		if index == len(value) || value[index] != '0' && value[index] != '1' {
			return false
		}
	}
	return true
}

func (RequestPath) String() string                      { return valueRedaction }
func (RequestPath) GoString() string                    { return valueRedaction }
func (path RequestPath) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (RequestPath) MarshalJSON() ([]byte, error)        { return rejectDirectJSONMarshal() }
