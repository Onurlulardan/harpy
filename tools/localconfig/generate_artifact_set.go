package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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
	set := newArtifactSetFor(options)
	if options.profile == profileForgeGuard {
		actorID, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("generate documents billing actor ID: %w", err)
		}
		set.billingActorID = actorID.String()
	}
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
		addAIMCPRuntimeArtifacts,
		addProfileEnvironmentArtifacts,
		addForgeGuardInputArtifacts,
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
