package main

import (
	"fmt"
	"time"
)

func generateArtifactSet(now time.Time) (*artifactSet, error) {
	return generateArtifactSetFor(now, localGenerationOptions())
}

func generateArtifactSetFor(
	now time.Time,
	options generationOptions,
) (*artifactSet, error) {
	canonical := now.UTC().Truncate(time.Second)
	if canonical.IsZero() {
		return nil, fmt.Errorf("clock is required")
	}
	if options.profile != profileLocal {
		return nil, fmt.Errorf("generation profile is invalid")
	}
	set := newArtifactSetFor(options)
	steps := []func(*artifactSet, time.Time) error{
		addConnectionArtifacts,
		addGeneralKeyrings,
		addIAMKeyrings,
		addPasswordBlocklists,
		addActorBridgeArtifacts,
		addRuntimePKI,
		addRecoveryArtifacts,
		addWorkloadIdentityArtifacts,
		addAuthorityArtifacts,
		addProfileEnvironmentArtifacts,
	}
	for _, step := range steps {
		if err := step(set, canonical); err != nil {
			return nil, err
		}
	}
	if err := set.requireComplete(); err != nil {
		return nil, err
	}
	return set, nil
}
