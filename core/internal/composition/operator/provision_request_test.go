package operator

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/adapters/provisionlaunch"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
)

func TestShowProvisionRequestPrintsCanonicalTenantID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "provision-request.json")
	previous := provisionlaunch.RequestPath
	provisionlaunch.RequestPath = path
	t.Cleanup(func() { provisionlaunch.RequestPath = previous })
	const tenantID = "00112233-4455-6677-8899-aabbccddeeff"
	if err := provisionlaunch.Run(context.Background(), tenantID); !errors.Is(err, provisionlaunch.ErrPending) {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"show-provision-request"}, &stdout, &stderr)
	if code != processcode.Clean || stdout.String() != tenantID+"\n" || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestShowProvisionRequestRejectsMissingFile(t *testing.T) {
	previous := provisionlaunch.RequestPath
	provisionlaunch.RequestPath = filepath.Join(t.TempDir(), "provision-request.json")
	t.Cleanup(func() { provisionlaunch.RequestPath = previous })
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), []string{"show-provision-request"}, &stdout, &stderr)
	if code != processcode.InvalidStartup || stdout.Len() != 0 ||
		!strings.Contains(stderr.String(), "\"stage\":\"provision_request\"") ||
		!strings.Contains(stderr.String(), "\"exit_code\":2") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}
