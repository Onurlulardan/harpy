package compatibilityhttp

import (
	"context"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

type Application interface {
	AuthorizeLegacy(context.Context, compatibilityapp.LegacyKey) (compatibilityapp.LegacyMapping, error)
}

type PublicApplication interface {
	Execute(context.Context, publicationapp.PublicRequest) (publicationapp.PublicResult, error)
}
