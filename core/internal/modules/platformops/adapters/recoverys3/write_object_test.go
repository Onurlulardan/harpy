package recoverys3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestPutVerifiedRequiresSeparateVersionReadbackAndRetention(t *testing.T) {
	api := newObjectAPIFixture()
	body := []byte("encrypted recovery archive")
	digest := sha256.Sum256(body)
	writer := NewWriter(api, "harpy-recovery", "owner", 30, time.Now().UTC())
	evidence, err := writer.PutVerified(context.Background(), WriteRequest{
		Key:  "environments/e/recovery-sets/r/database.fcb1",
		Body: bytes.NewReader(body), ContentLength: int64(len(body)),
		ExpectedSHA256: hex.EncodeToString(digest[:]), ExpectedBytes: int64(len(body)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if api.puts != 1 || api.gets != 1 ||
		evidence.VersionID == "" || evidence.RetentionMode != "COMPLIANCE" ||
		evidence.UploadSHA256 != evidence.ReadbackSHA256 ||
		evidence.UploadedBytes != evidence.ReadbackBytes {
		t.Fatalf("incomplete immutable evidence: %#v", evidence)
	}
}
