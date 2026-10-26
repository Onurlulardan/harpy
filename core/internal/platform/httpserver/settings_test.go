package httpserver

import (
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

func TestPublicServerUsesStreamingSafeGlobalTimeouts(t *testing.T) {
	loaded, err := loadTestAPIConfig(t)
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewPublic(
		loaded.PublicListenAddress(),
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		loaded.HTTP(),
		log.New(io.Discard, "", 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if server.ReadTimeout != 0 || server.WriteTimeout != 0 ||
		server.ReadHeaderTimeout != 5*time.Second || server.IdleTimeout != time.Minute ||
		server.MaxHeaderBytes != 65536 {
		t.Fatalf("unexpected public server settings: %#v", server)
	}
}

func TestManagementServerUsesFixedBounds(t *testing.T) {
	server, err := NewManagement(
		"127.0.0.1:9000",
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		log.New(io.Discard, "", 0),
	)
	if err != nil {
		t.Fatal(err)
	}
	if server.ReadHeaderTimeout != 2*time.Second || server.ReadTimeout != 5*time.Second ||
		server.WriteTimeout != 5*time.Second || server.IdleTimeout != 30*time.Second ||
		server.MaxHeaderBytes != 16384 {
		t.Fatalf("unexpected management server settings: %#v", server)
	}
}

func loadTestAPIConfig(t *testing.T) (config.API, error) {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", "postgresql://user@example.invalid/harpy?sslmode=disable")
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", httpServerTestCursorKeyring)
	configfixture.SetAPI(t)
	return config.LoadAPI()
}

const httpServerTestCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`
