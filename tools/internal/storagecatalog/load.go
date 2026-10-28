package storagecatalog

import (
	"fmt"

	"github.com/Onurlulardan/harpy/tools/internal/strictjson"
)

func Load(repositoryRoot string) (Manifest, error) {
	repository, err := openRepositoryRoot(repositoryRoot)
	if err != nil {
		return Manifest{}, err
	}
	defer repository.Close()
	rootContent, err := readBoundedFile(
		repository, rootDocumentPath, maximumRootBytes,
	)
	if err != nil {
		return Manifest{}, fmt.Errorf("read storage catalog root: %w", err)
	}
	var root rootDocument
	if err := strictjson.Decode(rootContent, &root); err != nil {
		return Manifest{}, fmt.Errorf("decode storage catalog root: %w", err)
	}
	if root.Version != catalogVersion {
		return Manifest{}, fmt.Errorf(
			"unsupported storage catalog version %d", root.Version,
		)
	}
	includeNames, err := validateIncludes(root.Includes)
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{
		Version: catalogVersion,
		Owners:  make([]Owner, 0, len(root.Includes)),
	}
	totalBytes := int64(len(rootContent))
	seenOwners := make(map[string]bool, len(root.Includes))
	fragmentIncludes := map[string]bool{}
	previousOwner := ""
	for index, include := range root.Includes {
		content, err := readBoundedFile(
			repository, include, maximumFragmentBytes,
		)
		if err != nil {
			return Manifest{}, fmt.Errorf("read storage owner fragment: %w", err)
		}
		totalBytes += int64(len(content))
		if totalBytes > maximumCatalogBytes {
			return Manifest{}, fmt.Errorf("storage catalog total byte limit exceeded")
		}
		owner, fragmentBytes, declared, err := loadOwner(
			repository, content, include, includeNames[include],
		)
		if err != nil {
			return Manifest{}, err
		}
		totalBytes += fragmentBytes
		if totalBytes > maximumCatalogBytes {
			return Manifest{}, fmt.Errorf("storage catalog total byte limit exceeded")
		}
		for _, fragment := range declared {
			if fragmentIncludes[fragment] {
				return Manifest{}, fmt.Errorf("duplicate storage owner fragment %q", fragment)
			}
			fragmentIncludes[fragment] = true
		}
		if err := validateOwner(owner, includeNames[include]); err != nil {
			return Manifest{}, err
		}
		if seenOwners[owner.Name] {
			return Manifest{}, fmt.Errorf("duplicate storage owner %q", owner.Name)
		}
		if index > 0 && previousOwner >= owner.Name {
			return Manifest{}, fmt.Errorf("storage owners must be sorted by name")
		}
		seenOwners[owner.Name] = true
		previousOwner = owner.Name
		manifest.Owners = append(manifest.Owners, owner)
	}
	if err := auditOwnerDirectory(repository, includeNames); err != nil {
		return Manifest{}, err
	}
	if err := auditFragmentDirectory(repository, fragmentIncludes); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}
