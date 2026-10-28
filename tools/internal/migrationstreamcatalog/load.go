package migrationstreamcatalog

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
	rootContent, err := readBoundedFile(repository, rootDocumentPath, maximumRootBytes)
	if err != nil {
		return Manifest{}, fmt.Errorf("read migration stream catalog root: %w", err)
	}
	var root rootDocument
	if err := strictjson.Decode(rootContent, &root); err != nil {
		return Manifest{}, fmt.Errorf("decode migration stream catalog root: %w", err)
	}
	if root.Version != catalogVersion {
		return Manifest{}, fmt.Errorf(
			"unsupported migration stream catalog version %d", root.Version,
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
	previousOwner := ""
	for index, include := range root.Includes {
		content, err := readBoundedFile(repository, include, maximumFragmentBytes)
		if err != nil {
			return Manifest{}, fmt.Errorf("read migration stream fragment: %w", err)
		}
		totalBytes += int64(len(content))
		if totalBytes > maximumCatalogBytes {
			return Manifest{}, fmt.Errorf("migration stream catalog total byte limit exceeded")
		}
		var document ownerDocument
		if err := strictjson.Decode(content, &document); err != nil {
			return Manifest{}, fmt.Errorf("decode migration stream fragment %q: %w", include, err)
		}
		if document.Version != ownerDocumentVersion {
			return Manifest{}, fmt.Errorf(
				"unsupported migration stream owner document version %d", document.Version,
			)
		}
		if err := validateOwner(document, includeNames[include]); err != nil {
			return Manifest{}, err
		}
		if seenOwners[document.Owner] {
			return Manifest{}, fmt.Errorf("duplicate migration stream owner %q", document.Owner)
		}
		if index > 0 && previousOwner >= document.Owner {
			return Manifest{}, fmt.Errorf("migration stream owners must be sorted by name")
		}
		seenOwners[document.Owner] = true
		previousOwner = document.Owner
		manifest.Owners = append(manifest.Owners, Owner{
			Name: document.Owner, Streams: append([]Stream(nil), document.Streams...),
		})
	}
	if err := validateGlobalRanges(manifest.Streams()); err != nil {
		return Manifest{}, err
	}
	if err := validateGenerationFrontier(manifest.Streams()); err != nil {
		return Manifest{}, err
	}
	if err := auditStreamDirectory(repository, includeNames); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func validateGlobalRanges(streams []Stream) error {
	for index := 1; index < len(streams); index++ {
		if streams[index-1].VersionTo >= streams[index].VersionFrom {
			return fmt.Errorf("migration stream version ranges overlap")
		}
	}
	return nil
}
