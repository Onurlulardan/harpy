package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFrontendCheckoutPresentRequiresPackageAndLock(t *testing.T) {
	root := t.TempDir()
	present, err := frontendCheckoutPresent(root)
	if err != nil || present {
		t.Fatalf("empty checkout present=%t err=%v", present, err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	present, err = frontendCheckoutPresent(root)
	if err != nil || present {
		t.Fatalf("partial checkout present=%t err=%v", present, err)
	}
	if err := os.WriteFile(filepath.Join(root, "yarn.lock"), []byte("# lock\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	present, err = frontendCheckoutPresent(root)
	if err != nil || !present {
		t.Fatalf("complete checkout present=%t err=%v", present, err)
	}
}

func TestStopWithoutCheckoutSucceeds(t *testing.T) {
	cfg := runtimeConfig{
		ReceiptPath:     filepath.Join(t.TempDir(), "receipt.json"),
		ReadyURL:        "http://127.0.0.1:1/",
		CheckoutPresent: false,
	}
	if err := runStop(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestParseCommandAcceptsOptionalStart(t *testing.T) {
	command, optional, ok := parseCommand([]string{"start", "--if-present"})
	if !ok || command != "start" || !optional {
		t.Fatalf("optional start = %q %t %t", command, optional, ok)
	}
	if _, _, ok = parseCommand([]string{"stop", "--if-present"}); ok {
		t.Fatal("stop accepted --if-present")
	}
	if _, _, ok = parseCommand(nil); ok {
		t.Fatal("empty command was accepted")
	}
}
