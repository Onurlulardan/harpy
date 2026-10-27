package platformreports

import _ "embed"

//go:embed sql/platform/resolve_transition_target.sql
var resolveTransitionTargetSQL string
