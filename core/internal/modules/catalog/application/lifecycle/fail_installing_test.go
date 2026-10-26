package lifecycle

import (
	"testing"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func TestShouldFailInstallingOnDurableInstallStop(t *testing.T) {
	operation := catalogdomain.Operation{
		Kind: catalogdomain.OperationInstall, State: catalogdomain.OperationBlocked,
	}
	if !shouldFailInstalling(operation, catalogdomain.InstallationInstalling) {
		t.Fatal("blocked install left INSTALLING in place")
	}
	operation.State = catalogdomain.OperationFailed
	if !shouldFailInstalling(operation, catalogdomain.InstallationInstalling) {
		t.Fatal("failed install left INSTALLING in place")
	}
	if shouldFailInstalling(operation, catalogdomain.InstallationActive) {
		t.Fatal("active installation was marked failed")
	}
	operation.Kind = catalogdomain.OperationPurge
	if shouldFailInstalling(operation, catalogdomain.InstallationInstalling) {
		t.Fatal("purge stop failed an install")
	}
}
