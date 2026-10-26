package postgres

const postgresApplicationNameLimit = 63

func cutoverAdoptionApplicationName(runID string) string {
	return "harpy-cutover/" + runID
}

func cutoverFinalizerApplicationName(runID string) string {
	return "harpy-finalizer/" + runID
}
