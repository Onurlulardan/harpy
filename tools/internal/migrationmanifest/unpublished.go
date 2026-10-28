package migrationmanifest

import (
	"fmt"
	"os"
	"path/filepath"
)

const unpublishedMarkerSuffix = ".unpublished"

func unpublishedMarkerPath(directory, digest string) string {
	return filepath.Join(directory, digest+unpublishedMarkerSuffix)
}

func WriteUnpublishedMarker(directory, digest string) error {
	if directory == "" || digest == "" {
		return fmt.Errorf("unpublished migration marker path is invalid")
	}
	return os.WriteFile(unpublishedMarkerPath(directory, digest), []byte(digest+"\n"), 0o600)
}

func MarkReleaseApplied(directory, digest string) error {
	if directory == "" || digest == "" {
		return fmt.Errorf("applied migration marker path is invalid")
	}
	err := os.Remove(unpublishedMarkerPath(directory, digest))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func retractUnpublishedMaximal(directory, digest string) error {
	marker := unpublishedMarkerPath(directory, digest)
	if _, err := os.Stat(marker); err != nil {
		return err
	}
	artifact := filepath.Join(directory, digest+immutableManifestSuffix)
	if err := os.Remove(artifact); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(marker); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
