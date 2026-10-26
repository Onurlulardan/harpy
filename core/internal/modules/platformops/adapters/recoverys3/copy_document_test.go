package recoverys3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestCopyDocumentVerifiesSourceBeforeCanonicalTargetAndReadback(t *testing.T) {
	sourceAPI, targetAPI := newObjectAPIFixture(), newObjectAPIFixture()
	body := []byte("FCDOC1-ciphertext")
	digest := sha256.Sum256(body)
	source := ObjectIdentity{
		Key: "tenant/source.fcdoc1", VersionID: "source-v1",
		SHA256: hex.EncodeToString(digest[:]), Bytes: int64(len(body)),
	}
	sourceAPI.objects[objectKey("documents", source.Key, source.VersionID)] =
		storedObject{body: body, version: source.VersionID}
	evidence, err := CopyDocumentVersion(
		context.Background(),
		NewReader(sourceAPI, "documents", "owner"),
		NewWriter(targetAPI, "harpy-recovery", "owner", 30, time.Now().UTC()),
		source,
		"environments/e/recovery-sets/r/documents/tenants/t/versions/v.fcdoc1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if sourceAPI.gets != 2 || targetAPI.puts != 1 || targetAPI.gets != 1 ||
		evidence.SHA256 != source.SHA256 || evidence.Bytes != source.Bytes {
		t.Fatalf("copy/readback evidence mismatch: %#v", evidence)
	}
}

func TestCopyDocumentDoesNotWriteWhenSourceDigestDiffers(t *testing.T) {
	sourceAPI, targetAPI := newObjectAPIFixture(), newObjectAPIFixture()
	source := ObjectIdentity{
		Key: "tenant/source.fcdoc1", VersionID: "source-v1",
		SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Bytes:  3,
	}
	sourceAPI.objects[objectKey("documents", source.Key, source.VersionID)] =
		storedObject{body: []byte("bad"), version: source.VersionID}
	_, err := CopyDocumentVersion(context.Background(),
		NewReader(sourceAPI, "documents", "owner"),
		NewWriter(targetAPI, "harpy-recovery", "owner", 30, time.Now().UTC()),
		source, "target")
	if err == nil || targetAPI.puts != 0 {
		t.Fatal("source mismatch reached immutable target")
	}
}
