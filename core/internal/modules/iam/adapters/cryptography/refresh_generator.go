package cryptography

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
)

type RefreshGenerator struct{ generator refreshtoken.Generator }

func NewRefreshGenerator() RefreshGenerator {
	return RefreshGenerator{generator: refreshtoken.NewGenerator()}
}

func (provider RefreshGenerator) Generate() (authentication.IssuedRefresh, error) {
	return provider.generator.Generate()
}
