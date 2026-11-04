package foundation

import (
	"errors"
	"net/url"
	"strings"
)

type Config struct {
	AdminDSN       string
	Database       string
	GoosePassword  string
	APIPassword    string
	WorkerPassword string
}

func (cfg Config) Validate() error {
	if !ValidDatabaseName(cfg.Database) {
		return errors.New("database name is not a safe identifier")
	}
	parsed, err := url.Parse(cfg.AdminDSN)
	if err != nil || parsed.Scheme != "postgres" || parsed.Host == "" {
		return errors.New("admin DSN is invalid")
	}
	adminDatabase := strings.TrimPrefix(parsed.Path, "/")
	if adminDatabase == "" || adminDatabase == cfg.Database {
		return errors.New("admin DSN must target the maintenance database")
	}
	password, ok := parsed.User.Password()
	if !ok || password == "" {
		return errors.New("admin DSN is missing a password")
	}
	if !secretOK(cfg.GoosePassword) || !secretOK(cfg.APIPassword) ||
		!secretOK(cfg.WorkerPassword) {
		return errors.New("role password is missing or too short")
	}
	if cfg.GoosePassword == cfg.APIPassword ||
		cfg.GoosePassword == cfg.WorkerPassword ||
		cfg.APIPassword == cfg.WorkerPassword {
		return errors.New("role passwords must be distinct")
	}
	if cfg.GoosePassword == password || cfg.APIPassword == password ||
		cfg.WorkerPassword == password {
		return errors.New("role password must differ from the admin password")
	}
	return nil
}

func (cfg Config) Secrets() []string {
	values := []string{cfg.AdminDSN, cfg.GoosePassword, cfg.APIPassword, cfg.WorkerPassword}
	parsed, err := url.Parse(cfg.AdminDSN)
	if err != nil {
		return values
	}
	if password, ok := parsed.User.Password(); ok {
		values = append(values, password)
	}
	return values
}

func secretOK(value string) bool {
	return len(value) >= 16 && !strings.ContainsAny(value, " \t\r\n")
}
