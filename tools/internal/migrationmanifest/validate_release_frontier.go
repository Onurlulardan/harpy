package migrationmanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

const immutableManifestSuffix = ".migration-manifest.json"

type immutableRelease struct {
	digest   string
	manifest postgresv1.MigrationManifest
}

// ValidateReleaseFrontierAdvance derives release order from the immutable
// artifact with the largest entry frontier. Older historical branches cannot
// authorize a new release after a newer immutable frontier was published.
func ValidateReleaseFrontierAdvance(
	directory string,
	current postgresv1.MigrationManifest,
) error {
	return validateReleaseFrontierAdvance(directory, current, false)
}

func ValidateReleaseFrontierAdvanceForPublish(
	directory string,
	current postgresv1.MigrationManifest,
) error {
	return validateReleaseFrontierAdvance(directory, current, true)
}

func validateReleaseFrontierAdvance(
	directory string,
	current postgresv1.MigrationManifest,
	replaceUnpublished bool,
) error {
	if err := current.Validate(); err != nil {
		return fmt.Errorf("validate canonical migration frontier: %w", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read immutable migration releases: %w", err)
	}
	maximalLength := -1
	var maximalReleases []immutableRelease
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), immutableManifestSuffix) {
			continue
		}
		content, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return fmt.Errorf("read immutable migration release: %w", err)
		}
		artifact, err := attestation.VerifyMigrationManifestArtifact(content)
		if err != nil {
			return fmt.Errorf("verify immutable migration release: %w", err)
		}
		digest := strings.TrimSuffix(entry.Name(), immutableManifestSuffix)
		if digest != artifact.DigestSHA256() {
			return fmt.Errorf("immutable migration release filename differs from digest")
		}
		candidate := artifact.Value()
		switch {
		case len(candidate.Entries) > maximalLength:
			maximalLength = len(candidate.Entries)
			maximalReleases = []immutableRelease{{digest: digest, manifest: candidate}}
		case len(candidate.Entries) == maximalLength:
			maximalReleases = append(
				maximalReleases,
				immutableRelease{digest: digest, manifest: candidate},
			)
		}
	}
	if len(maximalReleases) == 0 {
		return nil
	}
	maximalRelease, err := selectLineageMaximalRelease(maximalReleases)
	if err != nil {
		return err
	}
	maximal := maximalRelease.manifest
	maximalDigest := maximalRelease.digest
	if !releaseManifestExtends(maximal, current) {
		if replaceUnpublished {
			if err := retractUnpublishedMaximal(directory, maximalDigest); err == nil {
				return validateReleaseFrontierAdvance(directory, current, true)
			}
		}
		return fmt.Errorf(
			"canonical migration frontier does not append to maximal immutable release %s",
			maximalDigest,
		)
	}
	currentContract := current
	currentContract.Entries = maximal.Entries
	if !reflect.DeepEqual(currentContract, maximal) {
		return fmt.Errorf(
			"canonical migration frontier contract differs from maximal immutable release %s",
			maximalDigest,
		)
	}
	if len(current.Entries) == len(maximal.Entries) {
		return nil
	}
	maximalGeneration :=
		maximal.Entries[len(maximal.Entries)-1].TargetDatabaseGeneration
	currentGeneration := current.Entries[len(current.Entries)-1].TargetDatabaseGeneration
	if currentGeneration <= maximalGeneration {
		return fmt.Errorf(
			"canonical migration frontier appends entries to immutable release %s without advancing target_database_generation (%d <= %d)",
			maximalDigest, currentGeneration, maximalGeneration,
		)
	}
	return nil
}
