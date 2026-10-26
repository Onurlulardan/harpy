package runtimev1

func snapshotWith(profile string, additions map[string]any) map[string]any {
	snapshot := profileSnapshot(profile)
	variables := snapshot["variables"].(map[string]any)
	for key, value := range additions {
		variables[key] = value
	}
	return snapshot
}
