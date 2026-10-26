package sqlidentifier

import (
	"errors"
	"strings"
)

// AppendSQL renders a validated token as one quoted PostgreSQL identifier.
func AppendSQL(token Token, builder *strings.Builder) error {
	if builder == nil {
		return errors.New("SQL identifier builder is nil")
	}
	if err := token.Validate(); err != nil {
		return err
	}
	builder.WriteByte('"')
	builder.WriteString(token.identifier)
	builder.WriteByte('"')
	return nil
}
