package sqlidentifier

import (
	"fmt"
	"io"
)

const redactedToken = "<sqlidentifier:redacted>"

func (Token) String() string {
	return redactedToken
}

func (Token) GoString() string {
	return redactedToken
}

func (Token) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, redactedToken)
}
