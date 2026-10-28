package storagecatalog

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLoadRejectsInvalidRootContracts(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"unsupported_version", `{"version":1,"includes":[]}`, "unsupported"},
		{"null_includes", `{"version":2,"includes":null}`, "must be an array"},
		{"unknown_member", `{"version":2,"includes":[],"extra":true}`, "exact schema key"},
		{"duplicate_key", `{"version":2,"version":2,"includes":[]}`, "duplicate key"},
		{"trailing_value", `{"version":2,"includes":[]} {}`, "trailing JSON value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeRawFile(t, root, rootDocumentPath, []byte(test.body))
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadRejectsInvalidOwnerContracts(t *testing.T) {
	tests := []struct {
		name string
		body func(t *testing.T) []byte
		want string
	}{
		{"unsupported_version", ownerBody(3, validOwner("iam")), "unsupported"},
		{"null_arrays", func(*testing.T) []byte {
			return []byte(`{"version":1,"owner":{"name":"iam"}}`)
		}, "arrays must not be null"},
		{"null_fk_columns", ownerBody(1, ownerWithNullForeignKey()), "foreign key arrays"},
		{"unknown_member", func(t *testing.T) []byte {
			content := ownerBody(1, validOwner("iam"))(t)
			return []byte(strings.TrimSuffix(string(content), "}}") + `,"extra":true}}`)
		}, "exact schema key"},
		{"trailing_value", func(t *testing.T) []byte {
			return append(ownerBody(1, validOwner("iam"))(t), []byte(` {}`)...)
		}, "trailing JSON value"},
		{"name_mismatch", ownerBody(1, validOwner("other")), "differs from fragment"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			include := ownerPath("iam")
			writeRoot(t, root, catalogVersion, []string{include})
			writeRawFile(t, root, include, test.body(t))
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load error = %v, want %q", err, test.want)
			}
		})
	}
}

func ownerBody(version int, owner Owner) func(*testing.T) []byte {
	return func(t *testing.T) []byte {
		t.Helper()
		content, err := json.Marshal(ownerDocument{Version: version, Owner: owner})
		if err != nil {
			t.Fatal(err)
		}
		return content
	}
}

func ownerWithNullForeignKey() Owner {
	owner := validOwner("iam")
	owner.ForeignKeys = []ForeignKey{{}}
	return owner
}
