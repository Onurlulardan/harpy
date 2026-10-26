package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

var errRawMemberInvalid = errors.New("REST Query raw member is invalid")

type RawMember struct {
	value   []byte
	present bool
}

func NewRawMember(value []byte) (RawMember, error) {
	if err := validateRawMemberValue(value); err != nil {
		return RawMember{}, err
	}
	return RawMember{value: cloneBytes(value), present: true}, nil
}

func (member RawMember) Value() ([]byte, bool) {
	if !member.present {
		return nil, false
	}
	return cloneBytes(member.value), true
}

func (member RawMember) Validate() error {
	if !member.present {
		if member.value != nil {
			return errRawMemberInvalid
		}
		return nil
	}
	return validateRawMemberValue(member.value)
}

func (member RawMember) clone() RawMember {
	if !member.present {
		return RawMember{}
	}
	return RawMember{value: cloneBytes(member.value), present: true}
}

func validateRawMemberValue(value []byte) error {
	if len(value) == 0 || !utf8.Valid(value) {
		return errRawMemberInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(value))
	var decoded json.RawMessage
	if err := decoder.Decode(&decoded); err != nil {
		return errRawMemberInvalid
	}
	if err := decoder.Decode(&decoded); !errors.Is(err, io.EOF) {
		return errRawMemberInvalid
	}
	return nil
}

func cloneBytes(value []byte) []byte {
	return append([]byte(nil), value...)
}
