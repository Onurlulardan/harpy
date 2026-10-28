package main

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveManagementListenAddress(t *testing.T) {
	if localManagementListenAddress != "127.0.0.1:9090" {
		t.Fatalf("local default constant = %s", localManagementListenAddress)
	}
	got, err := resolveManagementListenAddress("local", "", "")
	if err != nil || got != "127.0.0.1:9090" {
		t.Fatalf("local default = %s, %v", got, err)
	}
	got, err = resolveManagementListenAddress("local", "", "0.0.0.0:9090")
	if err != nil || got != "0.0.0.0:9090" {
		t.Fatalf("preserved = %s, %v", got, err)
	}
	got, err = resolveManagementListenAddress("local", "0.0.0.0:9090", "127.0.0.1:9090")
	if err != nil || got != "0.0.0.0:9090" {
		t.Fatalf("override = %s, %v", got, err)
	}
	if _, err = resolveManagementListenAddress("container", "", ""); err == nil {
		t.Fatal("container deployment defaulted a bind address")
	}
	got, err = resolveManagementListenAddress("container", containerManagementListenAddress, "")
	if err != nil || got != "0.0.0.0:9000" {
		t.Fatalf("container override = %s, %v", got, err)
	}
	if _, err = resolveManagementListenAddress("local", "127.0.0.1", ""); err == nil {
		t.Fatal("host without a port was accepted")
	}
	if _, err = resolveManagementListenAddress("satellite", "", ""); err == nil {
		t.Fatal("unknown deployment was accepted")
	}
}

func TestSetupManagementBindStaysLoopbackUntilOverridden(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults"}, "")
	env := readEnv(t, root, "single-tenant")
	if env["HARPY_MANAGEMENT_LISTEN_ADDR"] != "127.0.0.1:9090" ||
		env["HARPY_PUBLIC_LISTEN_ADDR"] != "127.0.0.1:8080" {
		t.Fatalf("listen = management %s public %s",
			env["HARPY_MANAGEMENT_LISTEN_ADDR"], env["HARPY_PUBLIC_LISTEN_ADDR"])
	}
	runOK(t, root, []string{"--defaults"}, "")
	if readEnv(t, root, "single-tenant")["HARPY_MANAGEMENT_LISTEN_ADDR"] != "127.0.0.1:9090" {
		t.Fatal("second setup changed the management bind")
	}
	runOK(t, root, []string{"--defaults", "--management-listen", "0.0.0.0:9090"}, "")
	if readEnv(t, root, "single-tenant")["HARPY_MANAGEMENT_LISTEN_ADDR"] != "0.0.0.0:9090" {
		t.Fatal("explicit management override was ignored")
	}
	runOK(t, root, []string{"--defaults"}, "")
	if readEnv(t, root, "single-tenant")["HARPY_MANAGEMENT_LISTEN_ADDR"] != "0.0.0.0:9090" {
		t.Fatal("second setup dropped the explicit management override")
	}
	err := run([]string{"--defaults", "--management-listen", "0.0.0.0"},
		strings.NewReader(""), io.Discard, io.Discard, root)
	if err == nil {
		t.Fatal("invalid management listen address was accepted")
	}
}

func TestContainerComposePublishesExplicitManagementBind(t *testing.T) {
	module, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	body := mustRead(t, filepath.Join(module, "ops", "local", "compose.runtime.yml"))
	if !strings.Contains(body, "HARPY_MANAGEMENT_LISTEN_ADDR: "+containerManagementListenAddress) {
		t.Fatal("container runtime lost the explicit non-loopback management bind")
	}
	if !strings.Contains(body, "http://127.0.0.1:9000/readyz") {
		t.Fatal("in-container readiness check no longer uses loopback")
	}
	if !strings.Contains(body, "127.0.0.1:${LOCAL_API_MANAGEMENT_PORT}:9000") {
		t.Fatal("host management publish is no longer loopback")
	}
}
