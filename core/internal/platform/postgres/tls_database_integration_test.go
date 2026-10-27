//go:build integration

package postgres

import (
	"context"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const postgresTLSIntegrationImage = "pgvector/pgvector:0.8.3-pg18-bookworm@sha256:" +
	"c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84"

type postgresTLSInstance struct {
	endpoint  *url.URL
	authority postgresTLSAuthority
	wrongCA   string
}

func startPostgresTLSInstance(
	t *testing.T,
	ctx context.Context,
) postgresTLSInstance {
	t.Helper()
	directory := t.TempDir()
	authority := writePostgresTLSAuthority(t, directory, "trusted")
	wrong := writePostgresTLSAuthority(t, directory, "wrong")
	configuration := filepath.Join(directory, "postgresql.conf")
	entrypoint := filepath.Join(directory, "postgres-tls-entrypoint.sh")
	content := []byte("listen_addresses='*'\nssl=on\n" +
		"ssl_ca_file='/tls/ca.pem'\n" +
		"ssl_cert_file='/tls/server.pem'\n" +
		"ssl_key_file='/tls/server-key.pem'\n")
	if err := os.WriteFile(configuration, content, 0o600); err != nil {
		t.Fatal(err)
	}
	script := []byte("#!/usr/bin/env bash\nset -Eeuo pipefail\n" +
		"chown postgres:postgres /tls/ca.pem /tls/server.pem /tls/server-key.pem\n" +
		"chmod 600 /tls/ca.pem /tls/server.pem /tls/server-key.pem\n" +
		"exec /usr/local/bin/docker-entrypoint.sh \"$@\"\n")
	if err := os.WriteFile(entrypoint, script, 0o700); err != nil {
		t.Fatal(err)
	}
	container, err := tcpostgres.Run(
		ctx, postgresTLSIntegrationImage,
		tcpostgres.WithConfigFile(configuration),
		tcpostgres.WithDatabase("harpy_tls"),
		tcpostgres.WithUsername("harpy_tls_runtime"),
		tcpostgres.WithPassword("tls-integration-only"),
		testcontainers.WithFiles(
			testcontainers.ContainerFile{HostFilePath: authority.caPath,
				ContainerFilePath: "/tls/ca.pem", FileMode: 0o600},
			testcontainers.ContainerFile{HostFilePath: authority.serverPath,
				ContainerFilePath: "/tls/server.pem", FileMode: 0o600},
			testcontainers.ContainerFile{HostFilePath: authority.keyPath,
				ContainerFilePath: "/tls/server-key.pem", FileMode: 0o600},
			testcontainers.ContainerFile{HostFilePath: entrypoint,
				ContainerFilePath: "/usr/local/bin/postgres-tls-entrypoint.sh", FileMode: 0o700},
		),
		testcontainers.WithEntrypoint("bash", "/usr/local/bin/postgres-tls-entrypoint.sh"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_INITDB_ARGS": "--encoding=UTF8 --locale=en_US.utf8 " +
				"--data-checksums --auth-local=trust --auth-host=scram-sha-256",
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		if container != nil {
			_ = testcontainers.TerminateContainer(container)
		}
		t.Fatalf("start exact TLS PostgreSQL 18.4 image: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate TLS PostgreSQL: %v", err)
		}
	})
	raw, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	endpoint.Host = net.JoinHostPort("localhost", endpoint.Port())
	return postgresTLSInstance{
		endpoint: endpoint, authority: authority, wrongCA: wrong.caPath,
	}
}

func (instance postgresTLSInstance) dsn(host, mode, root string) string {
	endpoint := *instance.endpoint
	endpoint.Host = net.JoinHostPort(host, instance.endpoint.Port())
	query := url.Values{"sslmode": []string{mode}}
	if root != "" {
		query.Set("sslrootcert", root)
	}
	endpoint.RawQuery = query.Encode()
	return endpoint.String()
}
