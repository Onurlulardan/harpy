package storagecatalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func validOwner(name string) Owner {
	return Owner{
		Name: name, Prefix: name + "_", Scopes: []string{},
		MigrationDirs: []DirectoryClaim{}, StaticSQLDirs: []DirectoryClaim{},
		PlatformSchemas: []string{}, SchemaUsageGrants: []SchemaGrant{},
		PlatformRoutines: []string{}, RoutineExecGrants: []RoutineGrant{},
		PlatformTables: []string{}, ExternalAccesses: []ExternalAccess{},
		ForeignKeys: []ForeignKey{}, LegacyTables: []LegacyTable{},
		DynamicNamespaces: []DynamicNamespaceClaim{},
	}
}

func ownerPath(name string) string {
	return ownerDirectoryPath + "/" + name + ".json"
}

func writeRoot(t *testing.T, root string, version int, includes []string) {
	t.Helper()
	writeJSONFile(t, root, rootDocumentPath, rootDocument{
		Version: version, Includes: includes,
	})
}

func writeOwner(t *testing.T, root string, version int, owner Owner) {
	t.Helper()
	writeJSONFile(t, root, ownerPath(owner.Name), ownerDocument{
		Version: version, Owner: owner,
	})
}

func writeJSONFile(t *testing.T, root, relative string, value any) {
	t.Helper()
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeRawFile(t, root, relative, content)
}

func writeRawFile(t *testing.T, root, relative string, content []byte) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, content, 0o600); err != nil {
		t.Fatal(err)
	}
}
