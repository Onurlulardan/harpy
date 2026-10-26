package migration

import "testing"

func TestValidateDocumentRejectsUnsafeSQLEncodingAndDirectives(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		mode    string
		hashOf  []byte
	}{
		{"tamper", []byte("-- +goose Up\nSELECT 2;\n"), "transactional", testFreshSQL},
		{"bom", append([]byte{0xef, 0xbb, 0xbf}, testFreshSQL...), "transactional", nil},
		{"crlf", []byte("-- +goose Up\r\nSELECT 1;\r\n"), "transactional", nil},
		{"not terminated", []byte("-- +goose Up\nSELECT 1;"), "transactional", nil},
		{"invalid utf8", []byte{'-', '-', ' ', '+', 'g', 'o', 'o', 's', 'e', ' ', 'U', 'p', '\n', 0xff, '\n'}, "transactional", nil},
		{"down", []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\n"), "transactional", nil},
		{"unknown directive", []byte("-- +goose Up\n-- +goose Temporary\nSELECT 1;\n"), "transactional", nil},
		{"unbalanced block", []byte("-- +goose Up\n-- +goose StatementBegin\nSELECT 1;\n"), "transactional", nil},
		{"transactional escape", []byte("-- +goose NO TRANSACTION\n-- +goose Up\nSELECT 1;\n"), "transactional", nil},
		{"non transactional missing directive", testFreshSQL, "non_transactional", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document := mustDocument(t, testOwner, testFreshPath, test.content)
			hashContent := test.content
			if test.hashOf != nil {
				hashContent = test.hashOf
			}
			expected := expectedDocument{
				owner: testOwner, scope: "platform", transactionMode: test.mode,
				sha256: testSQLDigest(hashContent),
			}
			if err := validateDocument(document, expected); err == nil {
				t.Fatal("unsafe migration SQL was accepted")
			}
		})
	}
}

func TestValidateDocumentAcceptsExactNonTransactionalDirective(t *testing.T) {
	content := []byte("-- +goose NO TRANSACTION\n-- +goose Up\nCREATE INDEX CONCURRENTLY x ON y (z);\n")
	document := mustDocument(t, testOwner, testFreshPath, content)
	expected := expectedDocument{
		owner: testOwner, scope: "platform",
		transactionMode: "non_transactional", sha256: testSQLDigest(content),
	}
	if err := validateDocument(document, expected); err != nil {
		t.Fatal(err)
	}
}

func TestInspectDocumentDoesNotTreatSQLTextAsGooseDirective(t *testing.T) {
	content := []byte("-- +goose Up\nSELECT '-- +goose Down';\n")
	document := mustDocument(t, testOwner, testFreshPath, content)
	metadata, err := InspectDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.TransactionMode != "transactional" || metadata.SHA256 != testSQLDigest(content) {
		t.Fatal("document metadata differs from exact raw SQL")
	}
}
