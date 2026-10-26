package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var authorityPostgresEnvironment = []string{
	"PGHOST", "PGPORT", "PGDATABASE", "PGUSER", "PGPASSWORD", "PGPASSFILE",
	"PGAPPNAME", "PGCONNECT_TIMEOUT", "PGSSLMODE", "PGSSLKEY", "PGSSLCERT",
	"PGSSLSNI", "PGSSLROOTCERT", "PGSSLPASSWORD", "PGSSLNEGOTIATION",
	"PGTARGETSESSIONATTRS", "PGSERVICE", "PGSERVICEFILE", "PGTZ", "PGOPTIONS",
	"PGMINPROTOCOLVERSION", "PGMAXPROTOCOLVERSION", "PGCHANNELBINDING",
	"PGREQUIREAUTH",
}

var authorityDSNAllowedKeys = []string{
	"host", "port", "database", "user", "password", "connect_timeout",
	"sslmode", "sslkey", "sslcert", "sslrootcert", "sslpassword", "sslsni",
	"sslnegotiation", "min_protocol_version", "max_protocol_version",
	"channel_binding", "require_auth",
}

type declaredAuthorityDSN struct {
	parsed      *url.URL
	user        string
	host        string
	database    string
	password    string
	passwordSet bool
}

func openAuthorityDatabase(
	ctx context.Context, dsnPath, environment string,
) (*pgx.Conn, error) {
	return openProfileDatabase(ctx, dsnPath, environment, false)
}

func openFinalizerDatabase(
	ctx context.Context, dsnPath, environment string,
) (*pgx.Conn, error) {
	return openProfileDatabase(ctx, dsnPath, environment, true)
}

func openProfileDatabase(
	ctx context.Context, dsnPath, environment string, requirePassword bool,
) (*pgx.Conn, error) {
	secret, err := readSecret(dsnPath)
	if err != nil {
		return nil, err
	}
	config, configErr := parseAuthorityDatabaseConfig(
		string(secret), environment, requirePassword,
	)
	for index := range secret {
		secret[index] = 0
	}
	if configErr != nil {
		return nil, configErr
	}
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("authority database connection failed")
	}
	return connection, nil
}

func parseAuthorityDatabaseConfig(
	dsn, environment string, requirePassword bool,
) (*pgx.ConnConfig, error) {
	if ambientPostgresConfigPresent() {
		return nil, fmt.Errorf("ambient PostgreSQL configuration is forbidden")
	}
	declared, err := validateAuthorityDSN(dsn, requirePassword)
	if err != nil {
		return nil, fmt.Errorf("authority database connection config is invalid")
	}
	parseDSN := declared.parsed.String()
	if !declared.passwordSet {
		copyURL := *declared.parsed
		copyURL.User = url.UserPassword(declared.user, "pgpass-disabled")
		parseDSN = copyURL.String()
	}
	config, err := pgx.ParseConfigWithOptions(parseDSN, pgx.ParseConfigOptions{
		ParseConfigOptions: pgconn.ParseConfigOptions{
			ConnStringAllowedKeys: authorityDSNAllowedKeys,
		},
	})
	if !declared.passwordSet && config != nil {
		config.Password = ""
	}
	if err != nil || !matchesDeclaredAuthorityDSN(config, declared, environment) {
		return nil, fmt.Errorf("authority database connection config is invalid")
	}
	config.RuntimeParams["application_name"] = authorityApplicationName
	return config, nil
}

func validateAuthorityDSN(
	dsn string, requirePassword bool,
) (declaredAuthorityDSN, error) {
	if !strings.HasPrefix(dsn, "postgres://") &&
		!strings.HasPrefix(dsn, "postgresql://") {
		return declaredAuthorityDSN{}, fmt.Errorf("URL form is required")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || parsed.Opaque != "" || parsed.Fragment != "" || parsed.User == nil ||
		strings.Contains(parsed.Host, ",") {
		return declaredAuthorityDSN{}, fmt.Errorf("URL authority is invalid")
	}
	password, passwordSet := parsed.User.Password()
	user, host := parsed.User.Username(), parsed.Hostname()
	database := strings.TrimPrefix(parsed.Path, "/")
	if user == "" || host == "" || parsed.Path == database || database == "" ||
		strings.Contains(database, "/") || strings.ContainsAny(user+host+database+password, "\x00") ||
		(passwordSet && password == "") || (requirePassword && !passwordSet) {
		return declaredAuthorityDSN{}, fmt.Errorf("URL identity is incomplete")
	}
	for key, values := range parsed.Query() {
		if len(values) != 1 || values[0] == "" || isAuthorityIdentityQueryKey(key) {
			return declaredAuthorityDSN{}, fmt.Errorf("URL query is ambiguous")
		}
	}
	return declaredAuthorityDSN{parsed: parsed, user: user, host: host,
		database: database, password: password, passwordSet: passwordSet}, nil
}

func isAuthorityIdentityQueryKey(key string) bool {
	switch key {
	case "host", "port", "database", "dbname", "user", "password":
		return true
	default:
		return false
	}
}

func matchesDeclaredAuthorityDSN(
	config *pgx.ConnConfig, declared declaredAuthorityDSN, environment string,
) bool {
	if config == nil || config.User != declared.user || config.Host != declared.host ||
		config.Database != declared.database || len(config.Fallbacks) != 0 ||
		len(config.RuntimeParams) != 0 || !validAuthorityTLS(config.TLSConfig, environment) {
		return false
	}
	if declared.passwordSet {
		return config.Password == declared.password
	}
	return config.Password == ""
}

func ambientPostgresConfigPresent() bool {
	for _, name := range authorityPostgresEnvironment {
		if os.Getenv(name) != "" {
			return true
		}
	}
	return false
}

func validAuthorityTLS(config *tls.Config, environment string) bool {
	strict := config != nil && !config.InsecureSkipVerify && config.ServerName != ""
	if environment == "staging" || environment == "production" {
		return strict
	}
	return config == nil || strict
}
