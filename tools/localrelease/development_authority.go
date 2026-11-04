package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const developmentAuthorityDirectory = "/run/release/current"

func runArtifact(arguments []string) error {
	if len(arguments) != 0 {
		return fmt.Errorf("artifact accepts no arguments")
	}
	return publishDevelopmentBootstrapAuthority(
		developmentAuthorityDirectory, time.Now().UTC(),
	)
}

func publishDevelopmentBootstrapAuthority(directory string, now time.Time) error {
	directory = filepath.Clean(directory)
	if !filepath.IsAbs(directory) {
		return fmt.Errorf("bootstrap authority output must be absolute")
	}
	pair, err := buildDevelopmentBootstrapAuthority(now, releaseIDsForDirectory(directory))
	if err != nil {
		return err
	}
	info, err := os.Lstat(directory)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("inspect bootstrap authority output: %w", err)
	}
	if err == nil && !info.IsDir() {
		return fmt.Errorf("bootstrap authority output is not a directory")
	}
	output, err := beginOwnedOutput(directory, err == nil)
	if err != nil {
		return err
	}
	defer output.abort()
	if err := publishArtifact(output.staging, pair.document); err != nil {
		return err
	}
	if err := publishArtifact(output.staging, pair.signature); err != nil {
		return err
	}
	return output.commit()
}

func buildDevelopmentBootstrapAuthority(
	now time.Time, ids releaseIDs,
) (signedArtifactPair, error) {
	catalog, err := developmentAuthorityCatalog()
	if err != nil {
		return signedArtifactPair{}, err
	}
	signer, err := newSigningAuthority(now)
	if err != nil {
		return signedArtifactPair{}, err
	}
	return buildBootstrapAuthority(now, ids, catalog, signer)
}

func developmentAuthorityCatalog() (releaseCatalog, error) {
	var catalog releaseCatalog
	var err error
	catalog.roles, err = newArtifact("roles-manifest.json", rolesManifest())
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.privileges, err = newArtifact("privileges-manifest.json", privilegesManifest())
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.ownership, err = newArtifact(
		"ownership-manifest.json", ownershipManifest(uuid.NewString()),
	)
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.bootstrapOperation, err = newArtifact("bootstrap-operation.json",
		bootstrapOperation(
			catalog.roles.digest, catalog.privileges.digest, catalog.ownership.digest,
		))
	if err != nil {
		return releaseCatalog{}, err
	}
	catalog.cleanupOperation, err = newArtifact("cleanup-operation.json",
		cleanupOperation(catalog.roles.digest, catalog.privileges.digest))
	if err != nil {
		return releaseCatalog{}, err
	}
	return catalog, nil
}
