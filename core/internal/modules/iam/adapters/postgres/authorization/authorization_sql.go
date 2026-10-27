package postgres

import _ "embed"

//go:embed sql/tenant/authorize.sql
var authorizeSQL string

//go:embed sql/tenant/authorize_work.sql
var authorizeWorkSQL string

//go:embed sql/tenant/authorize_analytics.sql
var authorizeAnalyticsSQL string

//go:embed sql/tenant/authorize_documents.sql
var authorizeDocumentsSQL string

//go:embed sql/tenant/authorize_portal.sql
var authorizePortalSQL string
