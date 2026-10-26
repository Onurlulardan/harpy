package postgres

import _ "embed"

//go:embed sql/tenant/reference_evidence.sql
var referenceEvidenceSQL string
