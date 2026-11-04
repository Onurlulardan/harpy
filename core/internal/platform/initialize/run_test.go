package initialize

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
)

func TestRunRecordsExactlyFourStages(t *testing.T) {
	var calls []string
	result, err := Run(context.Background(), requestWith(func(name string) error {
		calls = append(calls, name)
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	assertStages(t, result, []stageExpect{
		{StageInitialize, true}, {StageFoundation, true},
		{StageGoose, true}, {StageApplication, true},
	})
	if strings.Join(calls, ",") != "preflight,foundation,goose,application" {
		t.Fatalf("calls = %v", calls)
	}
}

func TestFoundationFailureSkipsLaterStages(t *testing.T) {
	var calls []string
	_, err := Run(context.Background(), requestWith(func(name string) error {
		calls = append(calls, name)
		if name == "foundation" {
			return errors.New("foundation relation missing")
		}
		return nil
	}))
	if err == nil || !strings.Contains(err.Error(), "foundation relation missing") {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "preflight,foundation" {
		t.Fatalf("calls = %v", calls)
	}
}

func TestGooseFailureSkipsApplicationAndPreservesCause(t *testing.T) {
	cfg := testConfig()
	var calls []string
	_, err := Run(context.Background(), Request{
		Config:     cfg,
		Preflight:  func(context.Context, foundation.Config) error { return nil },
		Foundation: func(context.Context, foundation.Config) error { return nil },
		Goose: func(context.Context) error {
			calls = append(calls, "goose")
			return fmt.Errorf("SQLSTATE 42501 password %s", cfg.GoosePassword)
		},
		Application: func(context.Context) error {
			calls = append(calls, "application")
			return nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "SQLSTATE 42501") ||
		strings.Contains(err.Error(), cfg.GoosePassword) || len(calls) != 1 {
		t.Fatal(err)
	}
	var stage *StageError
	if !errors.As(err, &stage) || stage.ExitCode() != 5 || !stage.Retryable {
		t.Fatalf("stage error = %#v", stage)
	}
}

func TestGooseFailureCanBeRetriedWithoutCheckpoint(t *testing.T) {
	attempts := 0
	request := requestWith(func(name string) error {
		if name == "goose" {
			attempts++
			if attempts == 1 {
				return errors.New("SQLSTATE 40P01 deadlock")
			}
		}
		return nil
	})
	if _, err := Run(context.Background(), request); err == nil ||
		!strings.Contains(err.Error(), "SQLSTATE 40P01") {
		t.Fatal(err)
	}
	result, err := Run(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	assertStages(t, result, []stageExpect{
		{StageInitialize, true}, {StageFoundation, true},
		{StageGoose, true}, {StageApplication, true},
	})
	if attempts != 2 {
		t.Fatalf("goose attempts = %d", attempts)
	}
}

func TestDirtyPreflightDoesNotCallFoundation(t *testing.T) {
	cfg := testConfig()
	called := false
	_, err := Run(context.Background(), Request{
		Config: cfg,
		Preflight: func(context.Context, foundation.Config) error {
			return &foundation.Error{
				Op: "preflight", Database: cfg.Database, Retryable: false,
				Err: errors.New("database is not an empty foundation target: public.legacy"),
			}
		},
		Foundation: func(context.Context, foundation.Config) error {
			called = true
			return nil
		},
		Goose:       func(context.Context) error { called = true; return nil },
		Application: func(context.Context) error { called = true; return nil },
	})
	if err == nil || called || !strings.Contains(err.Error(), "public.legacy") {
		t.Fatalf("called=%t err=%v", called, err)
	}
}

type stageExpect struct {
	name string
	ok   bool
}

func assertStages(t *testing.T, result Result, want []stageExpect) {
	t.Helper()
	if len(result.Stages) != len(want) {
		t.Fatalf("stages = %#v", result.Stages)
	}
	for index, stage := range result.Stages {
		if stage.Name != want[index].name || stage.OK != want[index].ok {
			t.Fatalf("stage %d = %#v", index, stage)
		}
	}
}

func requestWith(step func(string) error) Request {
	return Request{
		Config: testConfig(),
		Preflight: func(context.Context, foundation.Config) error {
			return step("preflight")
		},
		Foundation: func(context.Context, foundation.Config) error {
			return step("foundation")
		},
		Goose: func(context.Context) error { return step("goose") },
		Application: func(context.Context) error {
			return step("application")
		},
	}
}

func testConfig() foundation.Config {
	return foundation.Config{
		AdminDSN:       "postgres://admin:foundation-admin-test-secret@127.0.0.1:5432/postgres?sslmode=disable",
		Database:       "harpy_lifecycle",
		GoosePassword:  "goose-lifecycle-test-secret",
		APIPassword:    "api-lifecycle-test-secret01",
		WorkerPassword: "worker-lifecycle-test-secret",
	}
}
