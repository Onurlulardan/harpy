package migrationsupersession

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

var (
	ownerPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	shaPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	basePattern  = regexp.MustCompile(
		`^[0-9]{14}_[a-z][a-z0-9]*(?:_[a-z0-9]+)*\.(fresh|v1_adopted)\.sql$`,
	)
)

func validateCatalog(catalog Catalog) error {
	if !validContractVersion(catalog.contractVersion) ||
		!ownerPattern.MatchString(catalog.owner) ||
		(catalog.scope != "platform" && catalog.scope != "tenant") ||
		!validSourceDirectory(catalog.sourceDirectory) ||
		len(catalog.entries) == 0 || len(catalog.entries) > maximumCatalogEntries {
		return fmt.Errorf("migration supersession catalog identity is invalid")
	}
	oldPaths := make(map[string]bool, len(catalog.entries))
	replacementPaths := make(map[string]bool, len(catalog.entries))
	chains := make(map[string]map[int]map[string]Entry, len(catalog.entries)/2)
	previous := ""
	for _, entry := range catalog.entries {
		identity, origin, level, err := validateEntry(
			catalog.contractVersion, catalog.sourceDirectory, entry,
		)
		if err != nil {
			return err
		}
		if entry.oldPath <= previous || oldPaths[entry.oldPath] ||
			replacementPaths[entry.replacementPath] {
			return fmt.Errorf("migration supersession entries are duplicate or unordered")
		}
		previous = entry.oldPath
		oldPaths[entry.oldPath] = true
		replacementPaths[entry.replacementPath] = true
		if chains[identity] == nil {
			chains[identity] = map[int]map[string]Entry{}
		}
		if chains[identity][level] == nil {
			chains[identity][level] = map[string]Entry{}
		}
		chains[identity][level][origin] = entry
	}
	return validateChains(catalog.contractVersion, chains)
}

func validSourceDirectory(directory string) bool {
	return fs.ValidPath(directory) && !strings.HasSuffix(directory, "/")
}

func validContractVersion(version string) bool {
	return contractDepth(version) != 0
}

func contractDepth(version string) int {
	switch version {
	case ContractVersionV1:
		return 1
	case ContractVersionV2:
		return 2
	case ContractVersionV3:
		return 3
	default:
		return 0
	}
}

func validateEntry(
	contractVersion, sourceDirectory string,
	entry Entry,
) (string, string, int, error) {
	if !fs.ValidPath(entry.oldPath) || !fs.ValidPath(entry.replacementPath) ||
		!shaPattern.MatchString(entry.oldSHA256) ||
		!shaPattern.MatchString(entry.replacementSHA256) ||
		path.Base(entry.oldPath) != path.Base(entry.replacementPath) ||
		!basePattern.MatchString(path.Base(entry.oldPath)) {
		return "", "", 0, fmt.Errorf("migration supersession entry is invalid")
	}
	origin := originFromBase(path.Base(entry.oldPath))
	level := 0
	replacementDirectory := ""
	switch path.Dir(entry.oldPath) {
	case path.Join(sourceDirectory, origin):
		level = 1
		replacementDirectory = path.Join(
			sourceDirectory, "supersessions", origin,
		)
	case path.Join(sourceDirectory, "supersessions", origin):
		level = 2
		replacementDirectory = path.Join(
			sourceDirectory, "supersessions", "v2", origin,
		)
	case path.Join(sourceDirectory, "supersessions", "v2", origin):
		level = 3
		replacementDirectory = path.Join(
			sourceDirectory, "supersessions", "v3", origin,
		)
	default:
		return "", "", 0, fmt.Errorf("migration supersession path binding is invalid")
	}
	if path.Dir(entry.replacementPath) != replacementDirectory ||
		level > contractDepth(contractVersion) {
		return "", "", 0, fmt.Errorf("migration supersession path binding is invalid")
	}
	identity := strings.TrimSuffix(path.Base(entry.oldPath), "."+origin+".sql")
	return identity, origin, level, nil
}

func originFromBase(base string) string {
	if strings.HasSuffix(base, ".fresh.sql") {
		return "fresh"
	}
	return "v1_adopted"
}
