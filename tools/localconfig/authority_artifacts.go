package main

import (
	"crypto/ed25519"
	"fmt"
	"time"
)

var authorityArtifactNames = []string{
	"local-authority-ack-signing-seed",
	"local-authority-release-signing-seed",
}

func addAuthorityArtifacts(set *artifactSet, _ time.Time) error {
	if set.options.profile != profileLocal {
		return nil
	}
	for _, name := range authorityArtifactNames {
		seed, err := randomBytes(ed25519.SeedSize)
		if err != nil {
			return fmt.Errorf("generate %s: %w", name, err)
		}
		if err := set.add(name, seed); err != nil {
			clear(seed)
			return err
		}
		clear(seed)
	}
	return nil
}
