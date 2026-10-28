package storagecatalog

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

func auditFragmentDirectory(repository *os.Root, declared map[string]bool) error {
	info, err := inspectRepositoryPath(repository, fragmentDirectoryPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) && len(declared) == 0 {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("storage fragment path must be a directory")
	}
	directory, err := openInspectedFile(repository, fragmentDirectoryPath, info)
	if err != nil {
		return err
	}
	defer directory.Close()
	entries, err := directory.ReadDir(-1)
	if err != nil {
		return fmt.Errorf("read storage fragment directory: %w", err)
	}
	found := make(map[string]bool, len(declared))
	for _, entry := range entries {
		owner := entry.Name()
		ownerPath := fragmentDirectoryPath + "/" + owner
		entryInfo, err := inspectFragmentOwnerDirectory(repository, entry, ownerPath)
		if err != nil {
			return err
		}
		if !hasDeclaredOwnerFragment(declared, ownerPath+"/") {
			return fmt.Errorf("orphan storage owner fragment directory %q", ownerPath)
		}
		if err := auditOwnerFragments(repository, owner, ownerPath, entryInfo, declared, found); err != nil {
			return err
		}
	}
	for include := range declared {
		if !found[include] {
			return fmt.Errorf("missing storage owner fragment %q", include)
		}
	}
	return verifyOpenFileStable(repository, fragmentDirectoryPath, directory, info)
}

func inspectFragmentOwnerDirectory(
	repository *os.Root, entry fs.DirEntry, ownerPath string,
) (os.FileInfo, error) {
	if entry.Type()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("storage owner fragment directory %q is a symbolic link", ownerPath)
	}
	info, err := inspectRepositoryPath(repository, ownerPath)
	if err != nil {
		return nil, fmt.Errorf("inspect storage owner fragment directory %q: %w", ownerPath, err)
	}
	if !info.IsDir() || !ownerFilenamePattern.MatchString(entry.Name()+".json") {
		return nil, fmt.Errorf("storage owner fragment directory %q is invalid", ownerPath)
	}
	return info, nil
}

func auditOwnerFragments(
	repository *os.Root,
	owner string,
	ownerPath string,
	info os.FileInfo,
	declared map[string]bool,
	found map[string]bool,
) error {
	directory, err := openInspectedFile(repository, ownerPath, info)
	if err != nil {
		return err
	}
	defer directory.Close()
	entries, err := directory.ReadDir(-1)
	if err != nil {
		return fmt.Errorf("read storage owner fragment directory %q: %w", ownerPath, err)
	}
	for _, entry := range entries {
		include := ownerPath + "/" + entry.Name()
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("storage owner fragment %q is a symbolic link", include)
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect storage owner fragment %q: %w", include, err)
		}
		if !entryInfo.Mode().IsRegular() {
			return fmt.Errorf("storage owner fragment %q is not a regular file", include)
		}
		if err := validateOwnerFragmentPath(owner, include); err != nil {
			return err
		}
		if !declared[include] {
			return fmt.Errorf("orphan storage owner fragment %q", include)
		}
		found[include] = true
	}
	return verifyOpenFileStable(repository, ownerPath, directory, info)
}

func hasDeclaredOwnerFragment(declared map[string]bool, prefix string) bool {
	for include := range declared {
		if strings.HasPrefix(include, prefix) {
			return true
		}
	}
	return false
}
