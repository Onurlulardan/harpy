//go:build integration

package systemtransferartifact

import (
	"bytes"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func TestMinIOIAMSystemTransferArtifactRuntime(t *testing.T) {
	fixture := newMinIORuntimeFixture(t)
	tenantID, actorID, sourceID := fixtureV7(t), fixtureV7(t), fixtureV7(t)
	importBody := []byte("email,active\none@example.test,true\n")
	importV1 := sealMinIOImport(t, fixture.runtime, tenantID, actorID, sourceID, importBody)
	importV2 := putMinIOTamper(t, fixture.client, fixture.imports,
		importV1.ObjectKey, []byte("hostile import replacement"))
	assertMinIOImportBytes(t, fixture.runtime, tenantID, actorID,
		sourceID, importV1, importBody)
	tampered := importV1
	tampered.HeaderSHA256[0] ^= 0xff
	if _, err := readMinIOImport(t, fixture.runtime, tenantID, actorID,
		sourceID, tampered); !errors.Is(err, errIntegrity) {
		t.Fatalf("tampered import error = %v", err)
	}
	tampered = importV1
	tampered.ObjectVersion = importV2
	if _, err := readMinIOImport(t, fixture.runtime, tenantID, actorID,
		sourceID, tampered); !errors.Is(err, errIntegrity) {
		t.Fatalf("replacement import error = %v", err)
	}
	deleted, err := fixture.runtime.DeleteExactImportSource(t.Context(),
		importV1.ObjectKey, importV1.ObjectVersion)
	if err != nil || deleted.DeleteResponseSHA256 == [32]byte{} ||
		deleted.AbsenceProbeSHA256 == [32]byte{} {
		t.Fatalf("delete exact = %#v, %v", deleted, err)
	}
	if _, err := readMinIOImport(t, fixture.runtime, tenantID, actorID,
		sourceID, importV1); !errors.Is(err, errIntegrity) {
		t.Fatalf("deleted version error = %v", err)
	}
	assertMinIOVersionExists(t, fixture.client, fixture.imports,
		importV1.ObjectKey, importV2)

	exportID := fixtureV7(t)
	exportBody := []byte("{\"id\":\"one\"}\n")
	exportV1 := sealMinIOExport(t, fixture.runtime, tenantID, exportID, exportBody)
	exportV2 := putMinIOTamper(t, fixture.client, fixture.exports,
		exportV1.ObjectKey, []byte("hostile export replacement"))
	assertMinIOExportBytes(t, fixture.runtime, tenantID, exportID, exportV1, exportBody)
	tamperedExport := exportV1
	tamperedExport.ObjectVersion = exportV2
	if _, err := readMinIOExport(t, fixture.runtime, tenantID, exportID,
		int64(len(exportBody)), tamperedExport); !errors.Is(err, errIntegrity) {
		t.Fatalf("replacement export error = %v", err)
	}
	listed, err := fixture.client.ListObjectVersions(t.Context(),
		&s3.ListObjectVersionsInput{Bucket: &fixture.exports, Prefix: &exportV1.ObjectKey})
	if err != nil || len(listed.Versions) != 2 || len(listed.DeleteMarkers) != 0 {
		t.Fatalf("export versions = %#v, %v", listed, err)
	}
	requireIAMMinIOPublicACLRejected(t, fixture)
}

func putMinIOTamper(t *testing.T, client *s3.Client, bucket, key string,
	body []byte,
) string {
	t.Helper()
	size := int64(len(body))
	output, err := client.PutObject(t.Context(), &s3.PutObjectInput{
		Bucket: &bucket, Key: &key, Body: bytes.NewReader(body), ContentLength: &size,
	})
	if err != nil || output == nil || output.VersionId == nil || *output.VersionId == "" {
		t.Fatalf("tamper upload = %#v, %v", output, err)
	}
	return *output.VersionId
}

func assertMinIOVersionExists(t *testing.T, client *s3.Client, bucket, key,
	version string,
) {
	t.Helper()
	if _, err := client.HeadObject(t.Context(), &s3.HeadObjectInput{
		Bucket: &bucket, Key: &key, VersionId: &version,
	}); err != nil {
		t.Fatalf("surviving version %q = %v", version, err)
	}
}

func assertMinIOImportBytes(t *testing.T, runtime *Runtime, tenantID, actorID,
	sourceID uuid.UUID, artifact systemtransfer.SourceArtifact, expected []byte,
) {
	t.Helper()
	value, err := readMinIOImport(t, runtime, tenantID, actorID, sourceID,
		artifact)
	if err != nil || !bytes.Equal(value, expected) {
		t.Fatalf("import plaintext = %q, %v", value, err)
	}
}

func assertMinIOExportBytes(t *testing.T, runtime *Runtime, tenantID,
	exportID uuid.UUID, metadata systemtransfer.ExportArtifactMetadata,
	expected []byte,
) {
	t.Helper()
	value, err := readMinIOExport(t, runtime, tenantID, exportID,
		int64(len(expected)), metadata)
	if err != nil || !bytes.Equal(value, expected) {
		t.Fatalf("export plaintext = %q, %v", value, err)
	}
}
