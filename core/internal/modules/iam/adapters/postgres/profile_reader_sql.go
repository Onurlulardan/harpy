package postgres

import _ "embed"

//go:embed sql/tenant/profile_historical_schema.sql
var profileHistoricalSchemaSQL string

//go:embed sql/tenant/profile_active_catalog.sql
var profileActiveCatalogSQL string

//go:embed sql/tenant/profile_historical_document.sql
var profileHistoricalDocumentSQL string
