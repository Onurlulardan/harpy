package migrationstreamcatalog

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func auditStreamDirectory(repository *os.Root, includes map[string]string) error {
	info, err := inspectRepositoryPath(repository, streamDirectoryPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) && len(includes) == 0 {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("migration stream fragment path must be a directory")
	}
	directory, err := openInspectedFile(repository, streamDirectoryPath, info)
	if err != nil {
		return err
	}
	defer directory.Close()
	entries, err := directory.ReadDir(-1)
	if err != nil {
		return fmt.Errorf("read migration stream fragment directory: %w", err)
	}
	found := make(map[string]bool, len(entries))
	for _, entry := range entries {
		include := streamDirectoryPath + "/" + entry.Name()
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("migration stream fragment %q is a symbolic link", include)
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect migration stream fragment %q: %w", include, err)
		}
		if !entryInfo.Mode().IsRegular() {
			return fmt.Errorf("migration stream fragment %q is not a regular file", include)
		}
		if _, err := ownerNameFromPath(include); err != nil {
			return err
		}
		if _, declared := includes[include]; !declared {
			return fmt.Errorf("orphan migration stream fragment %q", include)
		}
		found[include] = true
	}
	for include := range includes {
		if !found[include] {
			return fmt.Errorf("missing migration stream fragment %q", include)
		}
	}
	return verifyOpenFileStable(repository, streamDirectoryPath, directory, info)
}
