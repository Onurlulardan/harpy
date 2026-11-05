package operator

import (
	"io"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/provisionlaunch"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func showProvisionRequest(stdout io.Writer, stderr io.Writer) processcode.Code {
	tenantID, err := provisionlaunch.ReadTenantID()
	if err != nil {
		writeFailureCause(stderr, "provision_request", processcode.InvalidStartup, err.Error())
		return processcode.InvalidStartup
	}
	if _, err = io.WriteString(stdout, tenantID+"\n"); err != nil {
		writeFailureCause(stderr, "provision_request", processcode.InvalidStartup, err.Error())
		return processcode.InvalidStartup
	}
	return processcode.Clean
}
