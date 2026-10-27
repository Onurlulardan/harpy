package postgresv1

import "testing"

func TestInspectMigrationSourceRejectsEveryNonCanonicalGooseAnnotation(t *testing.T) {
	tests := map[string]string{
		"compact no transaction": "--+goose NO TRANSACTION\n-- +goose Up\nSELECT 1;\n",
		"extra comment markers":  "---- +goose NO TRANSACTION\n-- +goose Up\nSELECT 1;\n",
		"leading whitespace":     " -- +goose NO TRANSACTION\n-- +goose Up\nSELECT 1;\n",
		"env substitution":       "-- +goose Up\n-- +goose ENVSUB ON\nSELECT '${DANGEROUS}';\n",
		"lowercase command":      "-- +goose up\nSELECT 1;\n",
		"annotation in comment":  "-- text +goose Up\n-- +goose Up\nSELECT 1;\n",
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := InspectMigrationSource([]byte(content)); err == nil {
				t.Fatal("Goose-compatible non-canonical annotation was accepted")
			}
		})
	}
}

func TestInspectMigrationSourceAcceptsCanonicalStatementBlock(t *testing.T) {
	content := []byte("-- +goose Up\n-- +goose StatementBegin\n" +
		"CREATE FUNCTION x() RETURNS void LANGUAGE sql AS $$ SELECT 1; $$;\n" +
		"-- +goose StatementEnd\n")
	metadata, err := InspectMigrationSource(content)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.TransactionMode != "transactional" || len(metadata.SHA256) != 64 {
		t.Fatal("canonical migration source metadata is invalid")
	}
}

func TestInspectMigrationSourceRejectsGooseUnterminatedStatement(t *testing.T) {
	content := []byte("-- +goose Up\nCREATE TABLE unsafe (id uuid)\n")
	if _, err := InspectMigrationSource(content); err == nil {
		t.Fatal("Goose-unterminated migration SQL was accepted")
	}
}

func TestInspectMigrationSourceRejectsSecondStatementBlock(t *testing.T) {
	content := []byte("-- +goose Up\n-- +goose StatementBegin\nSELECT 1;\n" +
		"-- +goose StatementEnd\n-- +goose StatementBegin\nSELECT 2;\n" +
		"-- +goose StatementEnd\n")
	if _, err := InspectMigrationSource(content); err == nil {
		t.Fatal("a second Goose statement block was accepted")
	}
}

func TestInspectMigrationSourceAcceptsPairedFunctionsInOneBlock(t *testing.T) {
	content := []byte("-- +goose Up\n-- +goose StatementBegin\n" +
		"CREATE FUNCTION x() RETURNS void LANGUAGE sql AS $$ SELECT 1; $$;\n" +
		"CREATE OR REPLACE FUNCTION y() RETURNS void LANGUAGE sql AS $$ SELECT 2; $$;\n" +
		"-- +goose StatementEnd\nREVOKE ALL ON FUNCTION x() FROM PUBLIC;\n")
	if _, err := InspectMigrationSource(content); err != nil {
		t.Fatal(err)
	}
}
