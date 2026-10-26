package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func appendIdentifier(builder *strings.Builder, token sqlidentifier.Token) error {
	if builder == nil || sqlidentifier.AppendSQL(token, builder) != nil {
		return ErrContract
	}
	return nil
}
