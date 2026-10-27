package catalogprojection

import (
	"testing"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	"github.com/google/uuid"
)

func TestPendingReceiptCanonicalizesLocalNanosecondTime(t *testing.T) {
	now := time.Date(2026, 8, 19, 11, 37, 32, 123, time.Local)
	snapshot := Snapshot{
		ReceiptID: uuid.MustParse(testReceipt),
		PackageID: uuid.MustParse(testPackageID), PackageSlug: "hr",
		PackageVersion: "1.0.0", ReleaseID: uuid.MustParse(testReleaseID),
		Component: contract.Component{ID: testComponent},
		State: contract.ProvisionAvailable, CreatedAt: now, UpdatedAt: now,
	}
	receipt := receiptFromSnapshot(snapshot)
	if receipt.Ready || receipt.CompletedAt.Location() != time.UTC ||
		receipt.CompletedAt.Nanosecond()%1000 != 0 ||
		!receipt.RetryAt.Equal(receipt.CompletedAt.Add(time.Second)) {
		t.Fatalf("pending receipt was not canonical: %#v", receipt)
	}
}
