package postgres

import _ "embed"

//go:embed sql/issue_iam_v1_actor_assertion.sql
var issueIAMV1ActorAssertionSQL string

//go:embed sql/reconcile_iam_v1_actor_assertion_hits.sql
var reconcileIAMV1ActorAssertionHitsSQL string

//go:embed sql/purge_iam_v1_actor_assertions.sql
var purgeIAMV1ActorAssertionsSQL string
