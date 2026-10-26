package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const identityResult = `{"contract_version":"operator-identity-reconcile-result.v1",` +
	`"status":"completed","restore_operation_id":"22222222-2222-4222-8222-222222222222",` +
	`"environment_id":"11111111-1111-4111-8111-111111111111",` +
	`"tenant_id":"33333333-3333-4333-8333-333333333333",` +
	`"admission_state":"identity_reactivation_required",` +
	`"invalidation_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",` +
	`"custody_generation":2,` +
	`"custody_result_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`

func TestReactivationHandoffRecordsAndShowsExactOperatorBinding(t *testing.T) {
	directory := t.TempDir()
	createdAt := time.Date(2026, 7, 23, 12, 0, 0, 7, time.UTC)
	err := runRecordReactivation([]string{
		"-target-directory", directory,
		"-frontend-base-url", "http://localhost:5173",
	}, strings.NewReader(identityResult), func() time.Time { return createdAt })
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, reactivationHandoffName)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() ||
		runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("handoff mode = %v, %v", info, err)
	}
	var output bytes.Buffer
	if err = runShowReactivation([]string{"-source", path}, &output); err != nil {
		t.Fatal(err)
	}
	want := "http://localhost:5173/post-restore-reactivation?" +
		"environment_id=11111111-1111-4111-8111-111111111111&" +
		"restore_operation_id=22222222-2222-4222-8222-222222222222"
	if !strings.Contains(output.String(), want) {
		t.Fatalf("handoff output = %q", output.String())
	}
}

func TestReactivationHandoffRejectsUnsafeInput(t *testing.T) {
	directory := t.TempDir()
	clock := func() time.Time { return time.Now().UTC() }
	for _, test := range []struct {
		name, result, baseURL string
	}{
		{"trailing operator result", identityResult + `{}`, "http://localhost:5173"},
		{"credentialed frontend URL", identityResult, "https://owner:password@example.test"},
		{"frontend URL query", identityResult, "https://example.test?secret=value"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := runRecordReactivation([]string{
				"-target-directory", directory, "-frontend-base-url", test.baseURL,
			}, strings.NewReader(test.result), clock)
			if err == nil {
				t.Fatal("unsafe handoff input was accepted")
			}
		})
	}
}
