package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultsDoNotRotateOrLeak(t *testing.T) {
	root := t.TempDir()
	first := runOK(t, root, []string{"--defaults"}, "")
	env := readEnv(t, root, "single-tenant")
	assertRustFS(t, env)
	if env["TENANCY_MODE"] != "single" || env["S3_PATH_STYLE"] != "true" || env["S3_PROVIDER"] != "rustfs" {
		t.Fatalf("defaults: %#v", env)
	}
	if env["S3_ENDPOINT"] != "http://127.0.0.1:9000" || env["S3_ACCESS_KEY"] != env["RUSTFS_ACCESS_KEY"] {
		t.Fatalf("rustfs endpoint: %#v", env)
	}
	assertCommands(t, first, "single-tenant")
	assertNoLeak(t, first, env)
	assertSecretFiles(t, root, "single-tenant", env)
	_ = runOK(t, root, []string{"--defaults"}, "")
	if again := readEnv(t, root, "single-tenant"); again["POSTGRES_PASSWORD"] != env["POSTGRES_PASSWORD"] || again["RUSTFS_ACCESS_KEY"] != env["RUSTFS_ACCESS_KEY"] || again["RABBITMQ_PASSWORD"] != env["RABBITMQ_PASSWORD"] {
		t.Fatal("second run rotated secrets")
	}
	rotated := runOK(t, root, []string{"--defaults", "--rotate-secrets"}, "")
	changed := readEnv(t, root, "single-tenant")
	if changed["POSTGRES_PASSWORD"] == env["POSTGRES_PASSWORD"] || changed["RUSTFS_ACCESS_KEY"] == env["RUSTFS_ACCESS_KEY"] {
		t.Fatal("rotate-secrets kept generated secrets")
	}
	assertRustFS(t, changed)
	assertNoLeak(t, rotated, changed)
	assertSecretFiles(t, root, "single-tenant", changed)
}

func TestProfilesStaySeparate(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults"}, "")
	single := readEnv(t, root, "single-tenant")
	out := runOK(t, root, []string{"--defaults", "--profile", "multi-tenant"}, "")
	multi := readEnv(t, root, "multi-tenant")
	if multi["TENANCY_MODE"] != "multi" || multi["DEPLOYMENT_MODE"] != "saas" || multi["S3_PATH_STYLE"] != "true" || multi["POSTGRES_PASSWORD"] == "" {
		t.Fatalf("multi: %#v", multi)
	}
	if readEnv(t, root, "single-tenant")["POSTGRES_PASSWORD"] != single["POSTGRES_PASSWORD"] {
		t.Fatal("multi-tenant setup rewrote single-tenant secrets")
	}
	assertCommands(t, out, "multi-tenant")
}

func TestExternalPathStyleAndPreserve(t *testing.T) {
	root := t.TempDir()
	aws := "multi-tenant\n9090\n5433\nappdb\nappuser\nexternal\nhttps://s3.eu-west-1.amazonaws.com\naws\neu-west-1\nmybucket\nAKIAEXAMPLEKEY\nawssecretvalue\n"
	out := runOK(t, root, nil, aws)
	env := readEnv(t, root, "multi-tenant")
	if env["S3_PATH_STYLE"] != "false" || env["S3_PROVIDER"] != "aws" || env["HTTP_PORT"] != "9090" || env["POSTGRES_DB"] != "appdb" {
		t.Fatalf("aws: %#v", env)
	}
	if env["S3_ENDPOINT"] != "https://s3.eu-west-1.amazonaws.com" || env["S3_ACCESS_KEY"] != "AKIAEXAMPLEKEY" || env["RUSTFS_ACCESS_KEY"] == env["S3_ACCESS_KEY"] {
		t.Fatal("aws credentials mixed with rustfs")
	}
	assertRustFS(t, env)
	assertCommands(t, out, "multi-tenant")
	assertNoLeak(t, out, env)

	minio := "single-tenant\n8080\n5432\nharpy\nharpy\nexternal\nhttp://127.0.0.1:9100\nminio\nus-east-1\nbucketname\nminioaccess\nminiosecret\n"
	runOK(t, root, nil, minio)
	got := readEnv(t, root, "single-tenant")
	if got["S3_PATH_STYLE"] != "true" || got["S3_PROVIDER"] != "minio" || got["S3_ACCESS_KEY"] != "minioaccess" {
		t.Fatalf("minio: %#v", got)
	}
	if strings.Contains(mustRead(t, filepath.Join(root, "deploy", "single-tenant", ".env")), "MINIO_") {
		t.Fatal("minio env name written")
	}
	kept := "single-tenant\n8081\n5432\nharpy\nharpy\nexternal\nhttp://127.0.0.1:9100\nminio\nus-east-1\notherbucket\n"
	runOK(t, root, nil, kept)
	next := readEnv(t, root, "single-tenant")
	if next["S3_SECRET_KEY"] != "miniosecret" || next["POSTGRES_PASSWORD"] != got["POSTGRES_PASSWORD"] || next["HTTP_PORT"] != "8081" || next["S3_BUCKET"] != "otherbucket" {
		t.Fatalf("preserved: %#v", next)
	}
	runOK(t, root, []string{"--rotate-secrets"}, kept)
	spun := readEnv(t, root, "single-tenant")
	if spun["S3_SECRET_KEY"] != "miniosecret" || spun["POSTGRES_PASSWORD"] == next["POSTGRES_PASSWORD"] || spun["RUSTFS_ACCESS_KEY"] == next["RUSTFS_ACCESS_KEY"] {
		t.Fatal("external rotate changed the wrong secrets")
	}
	assertRustFS(t, spun)
}

func TestRejectsBadInput(t *testing.T) {
	root := t.TempDir()
	input := "nope\nsingle-tenant\n8080\n5432\nharpy\nharpy\nexternal\nnot-a-url\nhttp://127.0.0.1:9000\nminio\nus-east-1\nharpy\nrustfsadmin\nGOODKEY\nrustfsadmin\ngoodsecret\n"
	runOK(t, root, nil, input)
	env := readEnv(t, root, "single-tenant")
	if env["S3_ACCESS_KEY"] != "GOODKEY" || env["S3_SECRET_KEY"] != "goodsecret" || env["S3_PATH_STYLE"] != "true" {
		t.Fatalf("retry: %#v", env)
	}
	if strings.Contains(mustRead(t, filepath.Join(root, "deploy", "single-tenant", ".env")), "rustfsadmin") {
		t.Fatal("wrote rustfsadmin")
	}
	if err := run(nil, strings.NewReader(""), io.Discard, io.Discard, root); err == nil {
		t.Fatal("expected eof")
	}
	if err := run([]string{"extra"}, strings.NewReader(""), io.Discard, io.Discard, root); err == nil {
		t.Fatal("expected arguments error")
	}
	if err := run([]string{"--help"}, strings.NewReader(""), io.Discard, io.Discard, root); !errors.Is(err, errHelp) {
		t.Fatal(err)
	}
}

func TestPackageDoesNotStartProcesses(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		text := mustRead(t, entry.Name())
		if strings.Contains(text, "os/exec") || strings.Contains(text, "exec.Command") {
			t.Fatalf("%s starts a process", entry.Name())
		}
	}
}

func runOK(t *testing.T, root string, args []string, stdin string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	if err := run(args, strings.NewReader(stdin), &out, &errOut, root); err != nil {
		t.Fatalf("run: %v\n%s", err, errOut.String())
	}
	return out.String()
}

func readEnv(t *testing.T, root, profile string) map[string]string {
	t.Helper()
	return parseEnv(mustRead(t, filepath.Join(root, "deploy", profile, ".env")))
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func assertCommands(t *testing.T, out, profile string) {
	t.Helper()
	want := resultText(profile)
	if out != want {
		t.Fatalf("commands:\n%s", out)
	}
}

func assertNoLeak(t *testing.T, out string, env map[string]string) {
	t.Helper()
	for _, key := range []string{"POSTGRES_PASSWORD", "RABBITMQ_PASSWORD", "S3_ACCESS_KEY", "S3_SECRET_KEY", "RUSTFS_ACCESS_KEY", "RUSTFS_SECRET_KEY"} {
		if env[key] != "" && strings.Contains(out, env[key]) {
			t.Fatalf("stdout leaked %s", key)
		}
	}
}

func assertRustFS(t *testing.T, env map[string]string) {
	t.Helper()
	if err := validateAccessKey(env["RUSTFS_ACCESS_KEY"]); err != nil {
		t.Fatal(err)
	}
	if strings.EqualFold(env["RUSTFS_SECRET_KEY"], "rustfsadmin") {
		t.Fatal("default rustfs secret")
	}
}

func assertSecretFiles(t *testing.T, root, profile string, env map[string]string) {
	t.Helper()
	dir := filepath.Join(root, "deploy", profile, "secrets")
	if strings.TrimRight(mustRead(t, filepath.Join(dir, "rustfs_access_key")), "\r\n") != env["RUSTFS_ACCESS_KEY"] {
		t.Fatal("access key file drifted")
	}
	if strings.TrimRight(mustRead(t, filepath.Join(dir, "postgres_password")), "\r\n") != env["POSTGRES_PASSWORD"] {
		t.Fatal("postgres password file drifted")
	}
	if !strings.Contains(mustRead(t, filepath.Join(dir, "rabbitmq.conf")), "default_pass = "+env["RABBITMQ_PASSWORD"]) {
		t.Fatal("rabbitmq config drifted")
	}
}
