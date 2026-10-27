package licenseauthoritysource

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
	"github.com/google/uuid"
)

func TestWP7ConnectedAirGapAndDevelopmentSourcesStaySeparated(t *testing.T) {
	document := []byte(`{"schema":"harpy-license-envelope.v2"}`)
	credential := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
	request := licenseRequest(licenseauthority.DeliveryConnected)
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		incoming *http.Request,
	) {
		body, err := io.ReadAll(incoming.Body)
		if err != nil {
			t.Error(err)
		}
		if incoming.URL.Path != "/api/v2/license-snapshots/issue" ||
			incoming.Header.Get("Authorization") != "Bearer "+credential ||
			bytes.Contains(body, []byte(credential)) {
			t.Error("connected source crossed its credential boundary")
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(document)
	}))
	defer server.Close()
	connected, err := NewConnected(
		server.URL, []byte(credential), server.Client().Transport, time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, connected, request, document)

	path := filepath.Join(t.TempDir(), "license-snapshot.json")
	if err = os.WriteFile(path, document, 0o600); err != nil {
		t.Fatal(err)
	}
	fileSource, err := NewSnapshotFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, fileSource, licenseRequest(licenseauthority.DeliveryAirGap), document)
	assertSnapshot(t, fileSource, licenseRequest(licenseauthority.DeliveryDevelopment), document)
}

func assertSnapshot(
	t *testing.T,
	source licenseauthority.Source,
	request licenseauthority.IssueRequest,
	want []byte,
) {
	t.Helper()
	got, err := source.IssueSnapshot(context.Background(), request)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("snapshot = %s, %v", got, err)
	}
}

func licenseRequest(kind licenseauthority.DeliveryKind) licenseauthority.IssueRequest {
	return licenseauthority.IssueRequest{
		OperationID: uuid.New(),
		Binding: licenseauthority.Binding{
			LicenseID: uuid.New(), LicenseeOrganizationID: uuid.New(),
			EnvironmentID: uuid.New(), DeliveryKind: kind,
		},
		Nonce: [32]byte{1},
		ComponentVersions: []licenseauthority.ComponentVersion{
			{Component: "core-worker", Version: "v2.0.0"},
		},
	}
}
