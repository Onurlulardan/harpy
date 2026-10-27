package postrestore

import _ "embed"

//go:embed sql/platform/admission.sql
var admissionSQL string

//go:embed sql/platform/binding.sql
var bindingSQL string

//go:embed sql/platform/challenge.sql
var issueChallengeSQL string

//go:embed sql/platform/challenge_context.sql
var challengeContextSQL string

//go:embed sql/platform/fail_challenge.sql
var failChallengeSQL string

//go:embed sql/platform/reconcile.sql
var reconcilePlatformSQL string

//go:embed sql/platform/observe_reactivation.sql
var observeReactivationSQL string

//go:embed sql/tenant/owner_context.sql
var ownerContextSQL string

//go:embed sql/tenant/reactivation.sql
var completeSQL string

//go:embed sql/tenant/reconcile.sql
var reconcileTenantSQL string
