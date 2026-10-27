package licenseauthority

import "time"

type RuntimeEntitlements struct {
	State      RuntimeState
	Projection EntitlementProjection
}

func ProjectRuntimeEntitlements(
	snapshot VerifiedSnapshot,
	now time.Time,
) RuntimeEntitlements {
	projection := ProjectEntitlements(snapshot)
	state := snapshot.RuntimeState(now)
	if state != RuntimeBuildOff {
		return RuntimeEntitlements{State: state, Projection: projection}
	}
	for key := range projection.Boolean {
		if key != "workflow.enabled" && key != "api.access" {
			projection.Boolean[key] = false
		}
	}
	for key := range projection.Limits {
		projection.Limits[key] = Limit{}
	}
	return RuntimeEntitlements{State: state, Projection: projection}
}
