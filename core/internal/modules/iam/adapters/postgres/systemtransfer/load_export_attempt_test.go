package systemtransfer

import (
	"crypto/sha256"
	"strings"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func TestLoadExportAttemptPinsOwnedRoutineAndDecodesOwner(t *testing.T) {
	if !strings.Contains(outboundLoadSQL,
		"iam_system_record_export_worker_load_v1") {
		t.Fatal("export attempt loader does not pin the owned routine")
	}
	query := []byte(`{"select":["id"],"sort":[{"direction":"asc","field":"id"}]}`)
	generation := int64(2)
	raw := outboundAttemptRow{actorID: uuid.MustParse(
		"018f5e10-7b10-7000-8000-000000000002"), token: 3,
		authorization: 7, catalogVersionID: uuid.MustParse(
			"018f5e10-7b10-7000-8000-000000000003"),
		queryDocument: query, format: "jsonl", state: "running",
		attemptGeneration: &generation}
	catalog := sha256.Sum256([]byte("catalog-v1"))
	raw.catalogSHA = catalog[:]
	digest := sha256.Sum256(query)
	raw.querySHA = digest[:]
	claim := app.ExportJobClaim{TenantID: uuid.MustParse(
		"018f5e10-7b10-7000-8000-000000000001"), JobID: uuid.MustParse(
		"018f5e10-7b10-7000-8000-000000000004"), ExportID: uuid.MustParse(
		"018f5e10-7b10-7000-8000-000000000005"), Resource: publicresource.Users}
	attempt, err := raw.decode(claim)
	if err != nil || attempt.AttemptGeneration != generation {
		t.Fatalf("decode owner evidence: attempt=%+v err=%v", attempt, err)
	}
}
