package postgres

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestForcedPasswordCredentialCannotIssueNormalSession(t *testing.T) {
	const migration = "20260716013300_create_iam_login_credential_reader_v1"
	var reference []byte
	for _, origin := range []string{"fresh", "v1_adopted"} {
		body, err := os.ReadFile(filepath.Join(
			"..", "..", "migrations", "tenant", origin,
			migration+"."+origin+".sql",
		))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(body, []byte(
			"credential.state OPERATOR(pg_catalog.=) 'active'",
		)) {
			t.Fatal("normal login no longer requires an active password credential")
		}
		if bytes.Contains(body, []byte(
			"credential.state OPERATOR(pg_catalog.=) 'reset_required'",
		)) {
			t.Fatal("forced-password credential can issue a normal session")
		}
		if reference != nil && !bytes.Equal(reference, body) {
			t.Fatal("forced-password login gate differs between migration origins")
		}
		reference = body
	}
}
