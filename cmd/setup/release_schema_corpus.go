package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

type schemaSQL struct {
	owner, scope, path string
	body               []byte
}

type schemaSQLSet struct {
	owner, scope, root string
	globs              []string
	supersede          bool
}

func schemaMigrationSources(root string) ([]schemaSQL, error) {
	var documents []schemaSQL
	for _, set := range schemaSQLSets() {
		owned, err := readSchemaSQLSet(root, set)
		if err != nil {
			return nil, err
		}
		documents = append(documents, owned...)
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("fresh migration sources are absent")
	}
	return documents, nil
}

func schemaSQLSets() []schemaSQLSet {
	return []schemaSQLSet{
		{"platform_migrator", "platform", "core/internal/platform/migrator", []string{"migrations/fresh/*.sql"}, false},
		{"platform_cutover", "platform", "core/internal/platform/cutover", []string{"migrations/fresh/*.sql", "migrations/supersessions/fresh/*.sql"}, true},
		{"tenancy", "platform", "core/internal/modules/tenancy", []string{"migrations/platform/fresh/*.sql", "migrations/platform/supersessions/fresh/*.sql"}, false},
		{"tenancy", "tenant", "core/internal/modules/tenancy", []string{"migrations/tenant/fresh/*.sql"}, false},
		{"platformauthority", "platform", "core/internal/modules/platformauthority", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"platformops", "platform", "core/internal/modules/platformops", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"iam", "platform", "core/internal/modules/iam", []string{"migrations/platform/fresh/*.sql"}, false},
		{"iam", "tenant", "core/internal/modules/iam", []string{"migrations/tenant/fresh/*.sql", "migrations/tenant/supersessions/fresh/*.sql"}, false},
		{"system_resource_layout", "tenant", "core/internal/modules/system_resource_layout", []string{"migrations/tenant/fresh/*.fresh.sql", "migrations/tenant/supersessions/fresh/*.fresh.sql", "migrations/tenant/supersessions/v2/fresh/*.fresh.sql", "migrations/tenant/supersessions/v3/fresh/*.fresh.sql"}, false},
		{"billing", "platform", "core/internal/modules/billing", []string{"migrations/platform/fresh/*.sql"}, false},
		{"usage", "platform", "core/internal/modules/usage", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"platform_async", "platform", "core/internal/platform/async", []string{"migrations/fresh/*.sql", "migrations/supersessions/fresh/*.sql"}, true},
		{"platform_backupreceipt", "platform", "core/internal/platform/backupreceipt", []string{"migrations/fresh/*.sql"}, false},
		{"analytics", "platform", "core/internal/modules/analytics", []string{"migrations/platform/fresh/*.sql"}, false},
		{"analytics", "tenant", "core/internal/modules/analytics", []string{"migrations/tenant/fresh/*.sql", "migrations/tenant/supersessions/fresh/*.sql"}, false},
		{"catalog", "platform", "core/internal/modules/catalog", []string{"migrations/platform/fresh/*.sql"}, false},
		{"catalog", "tenant", "core/internal/modules/catalog", []string{"migrations/tenant/fresh/*.sql", "migrations/tenant/supersessions/fresh/*.sql"}, false},
		{"work", "tenant", "core/internal/modules/work", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
		{"documents", "tenant", "core/internal/modules/documents", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
		{"notifications", "platform", "core/internal/modules/notifications", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"notifications", "tenant", "core/internal/modules/notifications", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
		{"support", "platform", "core/internal/modules/support", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"audit", "tenant", "core/internal/modules/audit", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
	}
}

func readSchemaSQLSet(root string, set schemaSQLSet) ([]schemaSQL, error) {
	base := filepath.Join(root, filepath.FromSlash(set.root))
	seen := map[string]struct{}{}
	documents := make([]schemaSQL, 0)
	for _, glob := range set.globs {
		matches, err := filepath.Glob(filepath.Join(base, filepath.FromSlash(glob)))
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("fresh migration glob %s/%s is empty", set.root, glob)
		}
		for _, match := range matches {
			rel, err := filepath.Rel(base, match)
			if err != nil {
				return nil, err
			}
			logical := path.Join(set.root, filepath.ToSlash(rel))
			if _, exists := seen[logical]; exists {
				continue
			}
			seen[logical] = struct{}{}
			body, err := os.ReadFile(match)
			if err != nil {
				return nil, err
			}
			documents = append(documents, schemaSQL{
				owner: set.owner, scope: set.scope, path: logical, body: body,
			})
		}
	}
	if !set.supersede {
		return documents, nil
	}
	return dropSupersededSchemaSQL(root, set, documents)
}

func dropSupersededSchemaSQL(root string, set schemaSQLSet, documents []schemaSQL) ([]schemaSQL, error) {
	catalogPath := filepath.Join(root, filepath.FromSlash(set.root), "migration-supersessions.json")
	encoded, err := os.ReadFile(catalogPath)
	if err != nil {
		return nil, err
	}
	catalog, err := migrationsupersession.DecodeCatalog(encoded)
	if err != nil {
		return nil, err
	}
	indexed := make(map[string]schemaSQL, len(documents))
	for _, document := range documents {
		indexed[document.path] = document
	}
	superseded := map[string]bool{}
	replacements := map[string]bool{}
	for _, entry := range catalog.Entries() {
		if err = acceptSupersession(indexed, entry, superseded, replacements); err != nil {
			return nil, err
		}
	}
	reserved := set.root + "/migrations/supersessions/"
	selected := make([]schemaSQL, 0, len(documents))
	for _, document := range documents {
		if superseded[document.path] {
			continue
		}
		if strings.HasPrefix(document.path, reserved) && !replacements[document.path] {
			return nil, fmt.Errorf("fresh migration supersession %s is orphaned", document.path)
		}
		selected = append(selected, document)
	}
	return selected, nil
}

func acceptSupersession(
	indexed map[string]schemaSQL, entry migrationsupersession.Entry,
	superseded, replacements map[string]bool,
) error {
	if !strings.HasSuffix(entry.OldPath(), ".fresh.sql") {
		return nil
	}
	oldDocument, oldExists := indexed[entry.OldPath()]
	replacement, replacementExists := indexed[entry.ReplacementPath()]
	if !oldExists || !replacementExists ||
		schemaSQLDigest(oldDocument.body) != entry.OldSHA256() ||
		schemaSQLDigest(replacement.body) != entry.ReplacementSHA256() {
		return fmt.Errorf("fresh migration supersession %s differs", entry.OldPath())
	}
	superseded[entry.OldPath()] = true
	replacements[entry.ReplacementPath()] = true
	return nil
}

func schemaSQLDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return fmt.Sprintf("%x", sum)
}
