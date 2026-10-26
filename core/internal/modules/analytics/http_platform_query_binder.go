package analytics

import (
	querypostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/query"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func querypostgresBinder() queryapp.StoreBinder { return querypostgres.NewBinder() }
