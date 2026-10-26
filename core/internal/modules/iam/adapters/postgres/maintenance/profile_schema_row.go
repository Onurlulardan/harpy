package maintenance

import (
	"bytes"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type profileSchemaHeader struct {
	id            uuid.UUID
	resource      profileapp.ResourceKind
	version       int64
	state         string
	definitionSHA []byte
}

func (header profileSchemaHeader) same(other profileSchemaHeader) bool {
	return header.id == other.id && header.resource == other.resource &&
		header.version == other.version && header.state == other.state &&
		bytes.Equal(header.definitionSHA, other.definitionSHA)
}
