package main

import (
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestLoopbackListenAcceptsSameMachine(t *testing.T) {
	address, err := resolveManagementListenAddress("local", "", "")
	if err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		t.Fatalf("local management host = %q, %v", host, err)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	boundHost, _, err := net.SplitHostPort(listener.Addr().String())
	if err != nil || boundHost != host {
		t.Fatalf("bound address = %s, %v", listener.Addr(), err)
	}

	accepted := make(chan error, 1)
	go func() {
		accepted <- acceptLoopbackPing(listener, host)
	}()

	conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err = exchangeLoopbackPing(conn); err != nil {
		t.Fatal(err)
	}
	if err = <-accepted; err != nil {
		t.Fatal(err)
	}
}

func acceptLoopbackPing(listener net.Listener, host string) error {
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(time.Second))
	remoteHost, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	if remoteHost != host {
		return fmt.Errorf("accepted remote %s", remoteHost)
	}
	buf := make([]byte, 4)
	if _, err = io.ReadFull(conn, buf); err != nil {
		return err
	}
	if string(buf) != "ping" {
		return fmt.Errorf("unexpected payload")
	}
	_, err = conn.Write([]byte("pong"))
	return err
}

func exchangeLoopbackPing(conn net.Conn) error {
	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}
	if _, err := conn.Write([]byte("ping")); err != nil {
		return err
	}
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	if string(buf) != "pong" {
		return fmt.Errorf("unexpected reply")
	}
	return nil
}
