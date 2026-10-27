package backuparchive

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"io"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/backupcrypto"
)

type toolsFixture struct{ restored bytes.Buffer }

func (tools *toolsFixture) Dump(context.Context, string, string, io.Writer) error {
	return nil
}
func (tools *toolsFixture) Restore(_ context.Context, _ string, source io.Reader) error {
	_, err := io.Copy(&tools.restored, source)
	return err
}

type contentTools struct {
	toolsFixture
	content []byte
}

func (tools *contentTools) Dump(_ context.Context, _ string, _ string, target io.Writer) error {
	_, err := target.Write(tools.content)
	return err
}

func TestArchiveStreamsEncryptedDumpIntoRestore(t *testing.T) {
	private, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tools := &contentTools{content: []byte("custom-format-archive")}
	adapter := New(tools, tools)
	archiveContext := backupcrypto.Context{
		EnvironmentID: "018f0000-0000-7000-8000-000000000001",
		RecoverySetID: "018f0000-0000-7000-8000-000000000002",
		ComponentID:   "database_archive", DRKeyID: "dr_key_v1", RecoveryKeyID: "kit_key_v1"}
	drKey := bytes.Repeat([]byte{7}, 32)
	var encrypted bytes.Buffer
	_, err = adapter.Create(context.Background(), "dsn", "snapshot", &encrypted,
		archiveContext, backupcrypto.SealKeys{DRKEK: drKey,
			RecoveryPublic: private.PublicKey().Bytes(), Random: rand.Reader})
	if err != nil {
		t.Fatal(err)
	}
	_, err = adapter.Restore(context.Background(), "dsn", &encrypted, archiveContext,
		backupcrypto.OpenKeys{DRKEK: drKey})
	if err != nil || !bytes.Equal(tools.restored.Bytes(), tools.content) {
		t.Fatalf("restore mismatch: %v %q", err, tools.restored.Bytes())
	}
}
