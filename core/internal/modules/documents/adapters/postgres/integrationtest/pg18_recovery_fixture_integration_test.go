//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"
)

const (
	documentsRecoveryOwner       = "22222222-2222-4222-8222-222222222222"
	documentsRecoveryEnvironment = "018f0000-0000-7000-8000-000000000a10"
	documentsRecoverySet         = "018f0000-0000-7000-8000-000000000a11"
	documentsRecoveryFolder      = "018f0000-0000-7000-8000-000000000a12"
	documentsRecoveryItem        = "018f0000-0000-7000-8000-000000000a13"
	documentsRecoveryVersion     = "018f0000-0000-7000-8000-000000000a14"
	documentsRecoveryDeletion    = "018f0000-0000-7000-8000-000000000a15"
	documentsRecoveryRestore     = "018f0000-0000-7000-8000-000000000a16"
)

type documentsRecoveryFixture struct {
	schema, tenantID, objectKey, objectVersion, targetVersion string
	ciphertextSHA256, inventorySHA256                         [32]byte
	tenantSetSHA256                                           [32]byte
}

func newDocumentsRecoveryFixture(schema, tenantID string) documentsRecoveryFixture {
	return documentsRecoveryFixture{
		schema: schema, tenantID: tenantID,
		objectKey: "tenants/" + tenantID + "/documents/" +
			documentsRecoveryItem + "/versions/" +
			documentsRecoveryVersion + "/content",
		objectVersion:    "source-version-1",
		targetVersion:    "restored-version-2",
		ciphertextSHA256: sha256.Sum256([]byte(schema + ":ciphertext")),
		inventorySHA256:  sha256.Sum256([]byte(schema + ":inventory")),
		tenantSetSHA256:  sha256.Sum256([]byte(schema + ":tenant-set")),
	}
}

func prepareDocumentsRecoveryIdentity(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	createStatement := `CREATE TABLE ` + fixture.schema + `.tenancy_schema_identity(
 singleton boolean PRIMARY KEY,tenant_id uuid NOT NULL,
	 schema_name text NOT NULL,origin text NOT NULL)`
	if _, err := database.ExecContext(ctx, createStatement); err != nil {
		t.Fatal(err)
	}
	insertStatement := "INSERT INTO " + fixture.schema +
		".tenancy_schema_identity VALUES (true,$1,$2,$3)"
	if _, err := database.ExecContext(ctx, insertStatement,
		fixture.tenantID, fixture.schema, "fresh"); err != nil {
		t.Fatal(err)
	}
	grantStatement := "GRANT USAGE ON SCHEMA " + fixture.schema +
		" TO harpy_backup_reader,harpy_restore"
	if _, err := database.ExecContext(ctx, grantStatement); err != nil {
		t.Fatal(err)
	}
}

func seedDocumentsRecoveryVersion(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	plain := sha256.Sum256([]byte(fixture.schema + ":plaintext"))
	withDocumentsSchema(t, ctx, database, fixture.schema, func(tx *sql.Tx) {
		_, err := tx.ExecContext(ctx, `INSERT INTO documents_folders(
 id,parent_id,owner_id,display_name,lifecycle_state,row_version,
 created_at,updated_at,trashed_at)
 VALUES ($1,NULL,$2,'Recovery','ACTIVE',1,$3,$3,NULL)`,
			documentsRecoveryFolder, documentsRecoveryOwner, now)
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_items(
 id,folder_id,owner_id,current_version_id,lifecycle_state,row_version,
 created_at,updated_at)
 VALUES ($1,$2,$3,$4,'ACTIVE',1,$5,$5)`,
				documentsRecoveryItem, documentsRecoveryFolder, documentsRecoveryOwner,
				documentsRecoveryVersion, now)
		}
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_versions(
 id,item_id,version_number,lifecycle_state,original_name,declared_mime,
 detected_mime,file_kind,plaintext_bytes,ciphertext_bytes,plaintext_sha256,
 ciphertext_sha256,cipher_format,object_key,object_version,object_generation,
 key_id,wrapped_dek,scan_signature_revision,scanned_at,created_at)
 VALUES ($1,$2,1,'READY','recovery.pdf','application/pdf','application/pdf',
 'PDF',7,117,$3,$4,'FCDOC1',$5,$6,1,'documents-kek-1',$7,
 'clamav-recovery',$8,$8)`, documentsRecoveryVersion,
				documentsRecoveryItem, plain[:], fixture.ciphertextSHA256[:],
				fixture.objectKey, fixture.objectVersion, make([]byte, 115), now)
		}
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO documents_blob_deletions(
 id,target_kind,target_id,object_key,object_version,object_generation,
 plaintext_bytes,ciphertext_bytes,lifecycle_state,lease_generation,
 created_at,updated_at)
 VALUES ($1,'VERSION',$2,$3,$4,1,7,117,'PENDING',0,$5,$5)`,
				documentsRecoveryDeletion, documentsRecoveryVersion,
				fixture.objectKey, fixture.objectVersion, now)
		}
		if err != nil {
			t.Fatal(err)
		}
	})
}

func seedDocumentsRecoveryTransients(t *testing.T, ctx context.Context,
	database *sql.DB, fixture documentsRecoveryFixture,
) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	request := sha256.Sum256([]byte("recovery transient upload"))
	content := sha256.Sum256([]byte("recovery transient content"))
	seedDocumentsRecoveryUpload(t, ctx, database, fixture, now, request, content)
	seedDocumentsRecoveryArchives(t, ctx, database, fixture, now)
}
