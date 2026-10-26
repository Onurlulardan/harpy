package cataloghttp

import (
	"testing"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func TestRepresentInstallationIncludesLastOperation(t *testing.T) {
	id, _ := packagecontract.ParseStableID("018f1000-0000-7000-8000-000000000002")
	packageID, _ := packagecontract.ParsePackageID("018f1000-0000-7000-8000-000000000001")
	operationID, _ := packagecontract.ParseStableID("018f1000-0000-7000-8000-000000000003")
	value := catalogdomain.Installation{
		ID: id, PackageID: packageID, State: catalogdomain.InstallationFailed,
		ReferenceEpoch: 1, RowVersion: 2, LastOperationID: operationID,
	}
	wire := RepresentInstallation(value)
	if wire.LastOperationID == nil || *wire.LastOperationID != operationID.String() {
		t.Fatalf("last operation = %#v", wire.LastOperationID)
	}
	if RepresentInstallation(catalogdomain.Installation{
		ID: id, PackageID: packageID, State: catalogdomain.InstallationFailed,
		ReferenceEpoch: 1, RowVersion: 2,
	}).LastOperationID != nil {
		t.Fatal("missing last operation was presented")
	}
}
