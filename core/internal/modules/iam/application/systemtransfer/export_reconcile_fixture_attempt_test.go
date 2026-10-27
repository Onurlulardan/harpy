package systemtransfer

import (
	"crypto/sha256"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func fixtureExportAttemptWithoutTest(time.Time) ExportAttempt {
	query := []byte(`{"select":["id"],"sort":[{"direction":"asc","field":"id"}]}`)
	return ExportAttempt{TenantID: fixtureTenantID, ActorID: fixtureActorID,
		JobID: fixtureJobID, ExportID: fixtureTransferID,
		CatalogVersionID: fixtureExportCatalogID, Resource: publicresource.Users,
		Format: "jsonl", State: ExportWaiting, TokenVersion: 3,
		AuthorizationVersion: 7, CatalogSHA256: sha256.Sum256([]byte("catalog-v1")),
		QuerySHA256: sha256.Sum256(query), QueryDocument: query}
}
