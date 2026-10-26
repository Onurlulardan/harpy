package catalog

import catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"

func asyncUnavailable() error { return catalogapp.ErrUnavailable }
