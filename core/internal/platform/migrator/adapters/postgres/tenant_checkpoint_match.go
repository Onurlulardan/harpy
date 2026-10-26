package postgres

func tenantCheckpointMatches(
	checkpoint TenantCheckpoint,
	identity TenantCheckpointIdentity,
) bool {
	return checkpoint.TenantID == identity.TenantID &&
		checkpoint.Owner == identity.Owner && checkpoint.Version == identity.Version &&
		checkpoint.Origin == identity.Origin &&
		checkpoint.SelectedPath == identity.SelectedPath &&
		checkpoint.SelectedHash == identity.SelectedHash &&
		checkpoint.RunID == identity.RunID && checkpoint.Attempt == identity.Attempt
}

func TenantCheckpointMatchesSelection(
	checkpoint TenantCheckpoint,
	identity TenantCheckpointIdentity,
) bool {
	return checkpoint.TenantID == identity.TenantID &&
		checkpoint.Owner == identity.Owner && checkpoint.Version == identity.Version &&
		checkpoint.Origin == identity.Origin &&
		checkpoint.SelectedPath == identity.SelectedPath &&
		checkpoint.SelectedHash == identity.SelectedHash
}
