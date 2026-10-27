package contract

import (
	"errors"
	"fmt"
	"io"
)

const rawMemberRedaction = "<restquery.RawMember:redacted>"
const queryDocumentRedaction = "<restquery.QueryDocument:redacted>"

var errRawMemberJSON = errors.New("REST Query raw member cannot be encoded as JSON")
var errQueryDocumentJSON = errors.New("REST Query document cannot be encoded as JSON")

func (RawMember) String() string                     { return rawMemberRedaction }
func (RawMember) GoString() string                   { return rawMemberRedaction }
func (QueryDocument) String() string                 { return queryDocumentRedaction }
func (QueryDocument) GoString() string               { return queryDocumentRedaction }
func (RawMember) MarshalJSON() ([]byte, error)       { return nil, errRawMemberJSON }
func (QueryDocument) MarshalJSON() ([]byte, error)   { return nil, errQueryDocumentJSON }
func (RawMember) Format(state fmt.State, _ rune)     { writeRedaction(state, rawMemberRedaction) }
func (QueryDocument) Format(state fmt.State, _ rune) { writeRedaction(state, queryDocumentRedaction) }

func writeRedaction(state fmt.State, value string) {
	_, _ = io.WriteString(state, value)
}
