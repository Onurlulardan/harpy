//go:build integration

package migrator

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

type builtEntrypointProfile struct {
	name                  string
	binary                string
	arguments             []string
	dsn                   string
	service               runtimev1.Service
	sessionUser           string
	application           string
	api                   bool
	catalogPackageKeyring string
}

func runBuiltEntrypoint(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
	profile builtEntrypointProfile,
) {
	t.Helper()
	managementAddress := reserveBuiltEntrypointAddress(t)
	publicAddress := ""
	values := []string{
		"HARPY_ENVIRONMENT=test",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN=" + profile.dsn,
		"HARPY_POSTGRES_MAX_CONNS=2",
		"HARPY_POSTGRES_MIN_CONNS=0",
		"HARPY_POSTGRES_CONNECT_TIMEOUT=5s",
		"HARPY_POSTGRES_CLEANUP_TIMEOUT=5s",
		"HARPY_POSTGRES_HEALTH_CHECK_PERIOD=5s",
		"HARPY_READINESS_TIMEOUT=2s",
		"HARPY_MANAGEMENT_LISTEN_ADDR=" + managementAddress,
		"HARPY_DRAIN_TIMEOUT=5s",
		"HARPY_LOG_LEVEL=error",
	}
	if profile.api {
		publicAddress = reserveBuiltEntrypointAddress(t)
		values = append(
			values,
			"HARPY_PUBLIC_LISTEN_ADDR="+publicAddress,
			"HARPY_REST_QUERY_CURSOR_KEYRING="+builtEntrypointCursorKeyring,
		)
		values = appendBuiltEntrypointRuntimeFixture(
			values, configfixture.APIEnvironment(t, time.Now().UTC()),
			profile.catalogPackageKeyring,
		)
	} else if profile.service == runtimev1.ServiceCoreWorker {
		values = appendBuiltEntrypointRuntimeFixture(
			values, configfixture.WorkerEnvironment(t, time.Now().UTC()),
			profile.catalogPackageKeyring,
		)
	}
	if profile.catalogPackageKeyring != "" {
		values = append(values, "HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING="+
			profile.catalogPackageKeyring)
	}
	command := exec.CommandContext(ctx, profile.binary, profile.arguments...)
	command.Env = config.IntegrationSubprocessEnvironment(values)
	configureBuiltEntrypointProcess(command)
	logFile, err := os.CreateTemp(t.TempDir(), profile.name+"-*.log")
	if err != nil {
		t.Fatal(err)
	}
	command.Stdout, command.Stderr = logFile, logFile
	if err := command.Start(); err != nil {
		_ = logFile.Close()
		t.Fatalf("start built %s entrypoint: %v", profile.name, err)
	}
	done := make(chan error, 1)
	exited := make(chan struct{})
	go func() {
		done <- command.Wait()
		close(exited)
	}()
	finished := false
	defer func() {
		if !finished {
			_ = command.Process.Kill()
			<-done
		}
		_ = logFile.Close()
	}()
	waitBuiltEntrypointReady(
		t, managementAddress, profile.service, done, &finished, logFile,
	)
	assertBuiltEntrypointSession(t, ctx, database, profile)
	if profile.api {
		assertBuiltEntrypointEmptyAPI(t, publicAddress)
	}
	var heldDrainConnection net.Conn
	if profile.api {
		heldDrainConnection = openBuiltEntrypointDrainConnection(t, publicAddress)
		defer heldDrainConnection.Close()
	}
	if err := signalBuiltEntrypoint(command); err != nil {
		t.Fatalf("signal built %s entrypoint: %v", profile.name, err)
	}
	if profile.api {
		assertBuiltEntrypointDrainClosesListener(
			t, publicAddress, heldDrainConnection, exited,
		)
	}
	select {
	case err := <-done:
		finished = true
		if err != nil {
			t.Fatalf("built %s entrypoint did not drain cleanly: %v: %s",
				profile.name, err, readBuiltEntrypointLog(logFile))
		}
		if profile.api {
			assertBuiltEntrypointPublicClosed(t, publicAddress)
		}
	case <-time.After(15 * time.Second):
		t.Fatalf("built %s entrypoint did not stop after signal", profile.name)
	}
}

const builtEntrypointCursorKeyring = `{"version":1,"current":{"kid":"current-1","key":"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE"},"previous":{"kid":"previous-1","key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI"}}`

func reserveBuiltEntrypointAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func assertBuiltEntrypointSession(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
	profile builtEntrypointProfile,
) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var count int
		err := database.admin.QueryRow(ctx, `SELECT count(*) FROM pg_catalog.pg_stat_activity
WHERE datname=$1 AND usename=$2 AND application_name=$3`,
			builtEntrypointDatabaseName, profile.sessionUser, profile.application,
		).Scan(&count)
		if err == nil && count > 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("built %s entrypoint omitted exact PostgreSQL role session: %v",
				profile.name, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func readBuiltEntrypointLog(file *os.File) string {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err.Error()
	}
	document, err := io.ReadAll(io.LimitReader(file, 16*1024))
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("%s", document)
}
