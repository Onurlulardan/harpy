package goose

import (
	"os"
	"strings"
	"testing"
)

func TestPostRestoreAdmissionReturnsNoRowWhenIdle(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00809_iam_post_restore_admission_idle.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(body), "-- +goose Down")
	if !found || !strings.Contains(up, "WHERE false") {
		t.Fatal("idle post-restore admission still returns a row")
	}
}

func TestPostRestoreAdmissionStubStaysImmutable(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00290_iam_post_restore_admission_v1.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "WHERE false") {
		t.Fatal("applied post-restore admission migration changed")
	}
}
