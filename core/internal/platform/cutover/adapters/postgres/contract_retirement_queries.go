package postgres

import _ "embed"

//go:embed sql/admit_contract_retirement_manifest.sql
var admitContractRetirementManifestSQL string

//go:embed sql/record_contract_retirement_hits.sql
var recordContractRetirementHitsSQL string

//go:embed sql/provision_contract_retirement.sql
var provisionContractRetirementSQL string

//go:embed sql/reconcile_contract_retirement.sql
var reconcileContractRetirementSQL string

//go:embed sql/mark_contract_retirement_eligible.sql
var markContractRetirementEligibleSQL string

//go:embed sql/finalize_contract_retirement.sql
var finalizeContractRetirementSQL string
