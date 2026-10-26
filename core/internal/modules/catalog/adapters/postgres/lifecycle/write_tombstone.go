package lifecycle

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/write_tombstone.sql
var writeTombstoneSQL string

func (storage *store) writeTombstone(
	ctx context.Context,
	operation catalogdomain.Operation,
	installation catalogdomain.Installation,
	purgedAt time.Time,
) error {
	operationID, err := uuid.Parse(operation.ID.String())
	if err != nil || operation.Kind != catalogdomain.OperationPurge ||
		installation.State != catalogdomain.InstallationPurged {
		return unavailable()
	}
	receipts, err := storage.CompileOwnerReceipts(ctx, operationID)
	if err != nil || len(receipts) != len(operation.Plan.Components)*2 {
		return unavailable()
	}
	root := ownerReceiptRoot(receipts)
	var packageID, slug, packageSHA, storedRoot string
	var epoch int64
	if err = storage.queryRow(ctx, writeTombstoneSQL, installation.PackageID.String(),
		operation.Plan.Slug, operation.Plan.PackageSHA256, installation.ReferenceEpoch,
		root, purgedAt).Scan(&packageID, &slug, &packageSHA, &epoch, &storedRoot); err != nil ||
		packageID != installation.PackageID.String() || slug != operation.Plan.Slug ||
		packageSHA != operation.Plan.PackageSHA256 || epoch != int64(installation.ReferenceEpoch) ||
		storedRoot != root {
		return unavailable()
	}
	return nil
}

func ownerReceiptRoot(receipts []catalogdomain.OwnerReceipt) string {
	digests := make([]string, len(receipts))
	for index, receipt := range receipts {
		digests[index] = receiptDigest(receipt)
	}
	sort.Strings(digests)
	digest := sha256.Sum256([]byte(strings.Join(digests, "\x00")))
	return hex.EncodeToString(digest[:])
}
