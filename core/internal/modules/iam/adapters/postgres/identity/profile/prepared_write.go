package profilepostgres

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type preparedProfileWrite struct {
	expectedSourceSchemaID uuid.UUID
	sourceSchema           profileapp.Schema
	target                 profileapp.Document
}

type SelfPreparedWrite struct {
	ExpectedSourceSchemaID uuid.UUID
	SourceSchema           profileapp.Schema
	Target                 profileapp.Document
}
