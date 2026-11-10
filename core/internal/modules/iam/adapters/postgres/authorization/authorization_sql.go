package postgres

import _ "embed"

//go:embed sql/tenant/authorize.sql
var authorizeSQL string

//go:embed sql/tenant/authorize_documents.sql
var authorizeDocumentsSQL string
