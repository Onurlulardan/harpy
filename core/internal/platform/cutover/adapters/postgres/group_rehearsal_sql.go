package postgres

import _ "embed"

//go:embed sql/record_billing_rehearsal.sql
var recordBillingFormsRehearsalSQL string
