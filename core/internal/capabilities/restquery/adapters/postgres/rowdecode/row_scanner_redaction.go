package rowdecode

import (
	"fmt"
	"io"
	"log/slog"
)

func (RowScanner) String() string                         { return redacted }
func (RowScanner) GoString() string                       { return redacted }
func (RowScanner) LogValue() slog.Value                   { return slog.StringValue(redacted) }
func (scanner RowScanner) Format(state fmt.State, _ rune) { _, _ = io.WriteString(state, redacted) }
func (RowScanner) MarshalJSON() ([]byte, error)           { return nil, ErrDirectMarshal }
