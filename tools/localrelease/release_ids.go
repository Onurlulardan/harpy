package main

import "github.com/google/uuid"

type releaseIDs struct {
	release, environment, authorityManifest string
	clusterLease, extensionLease            string
	bootstrapRun, extensionRun, cleanupRun  string
	cutoverOperation, cutoverReceipt        string
}

func newReleaseIDs() releaseIDs {
	return newReleaseIDsForEnvironment("")
}

func newReleaseIDsForEnvironment(environment string) releaseIDs {
	if environment == "" {
		environment = uuid.Must(uuid.NewV7()).String()
	}
	return releaseIDs{
		release: uuid.Must(uuid.NewV7()).String(), environment: environment,
		authorityManifest: uuid.Must(uuid.NewV7()).String(),
		clusterLease:      uuid.Must(uuid.NewV7()).String(),
		extensionLease:    uuid.Must(uuid.NewV7()).String(),
		bootstrapRun:      uuid.Must(uuid.NewV7()).String(),
		extensionRun:      uuid.Must(uuid.NewV7()).String(),
		cleanupRun:        uuid.Must(uuid.NewV7()).String(),
		cutoverOperation:  uuid.Must(uuid.NewV7()).String(),
		cutoverReceipt:    uuid.Must(uuid.NewV7()).String(),
	}
}

func applyStoredReleaseID(ids releaseIDs, applied string) releaseIDs {
	if applied != "" {
		ids.release = applied
	}
	return ids
}
