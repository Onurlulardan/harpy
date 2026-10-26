//go:build integration

package postgres

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type builtCutoverProcess struct {
	command *exec.Cmd
	stdout  bytes.Buffer
	stderr  bytes.Buffer
	done    chan error
}

type builtCutoverOutcome struct {
	exitCode int
	stdout   []byte
	stderr   []byte
}

func buildCutoverMigratorBinary(t *testing.T, ctx context.Context) string {
	t.Helper()
	root := cutoverRepositoryRoot(t)
	name := "core-migrator"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	target := filepath.Join(t.TempDir(), name)
	command := exec.CommandContext(ctx, "go", "build", "-o", target, "./core/cmd/migrator")
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build cutover migrator: %v: %s", err, output)
	}
	return target
}

func cutoverRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve cutover test source")
	}
	directory := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("resolve backend-v2 repository root")
		}
		directory = parent
	}
}

func startBuiltCutoverProcess(
	t *testing.T, ctx context.Context, binary string, argument string, environment []string,
) *builtCutoverProcess {
	t.Helper()
	process := &builtCutoverProcess{done: make(chan error, 1)}
	process.command = exec.CommandContext(ctx, binary, argument)
	process.command.Env = config.IntegrationSubprocessEnvironment(environment)
	process.command.Stdout = &process.stdout
	process.command.Stderr = &process.stderr
	if err := process.command.Start(); err != nil {
		t.Fatalf("start built %s: %v", argument, err)
	}
	go func() { process.done <- process.command.Wait() }()
	return process
}

func waitBuiltCutoverProcess(
	t *testing.T, process *builtCutoverProcess, timeout time.Duration,
) builtCutoverOutcome {
	t.Helper()
	select {
	case err := <-process.done:
		code := 0
		if err != nil {
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) {
				t.Fatalf("wait built cutover process: %v", err)
			}
			code = exitError.ExitCode()
		}
		return builtCutoverOutcome{
			exitCode: code,
			stdout:   append([]byte(nil), process.stdout.Bytes()...),
			stderr:   append([]byte(nil), process.stderr.Bytes()...),
		}
	case <-time.After(timeout):
		_ = process.command.Process.Kill()
		<-process.done
		t.Fatal("built cutover process timed out")
		return builtCutoverOutcome{}
	}
}

func assertBuiltCutoverResult(
	t *testing.T, outcome builtCutoverOutcome, action string, expectedExit int,
) runtimev1.CutoverCommandResult {
	t.Helper()
	if outcome.exitCode != expectedExit {
		t.Fatalf("built %s exit = %d, stderr=%s", action, outcome.exitCode, outcome.stderr)
	}
	result, err := runtimev1.DecodeCutoverCommandResult(outcome.stdout)
	if err != nil {
		t.Fatalf("decode built %s stdout: %v: %s", action, err, outcome.stdout)
	}
	if result.Action != action || int(result.ExitCode) != outcome.exitCode {
		t.Fatalf("built %s process/result mismatch: %#v", action, result)
	}
	return result
}

func assertBuiltCutoverPreflightFailure(t *testing.T, outcome builtCutoverOutcome) {
	t.Helper()
	if outcome.exitCode != int(runtimev1.ExitCodeDependencyPreflight) || len(outcome.stdout) != 0 {
		t.Fatalf("cutover preflight outcome = %d stdout=%s stderr=%s",
			outcome.exitCode, outcome.stdout, outcome.stderr)
	}
}
