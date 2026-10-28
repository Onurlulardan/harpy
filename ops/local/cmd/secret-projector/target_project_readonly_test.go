package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestProjectTargetPublishesCompleteMarkerOnReadonlyDirectory(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	sourceRoot := writeTestSource(t, now)
	source, err := verifySource(sourceRoot, now)
	if err != nil {
		t.Fatal(err)
	}
	target := runtimeTarget("test", "local-ca.crt")
	if target.dir != 0o500 {
		t.Fatalf("target dir mode = %o, want 0500", target.dir)
	}
	wantPerm := directoryPermAfterChmod(t, target.dir)
	targetRoot := t.TempDir()
	keepTempDirRemovable(t, targetRoot)
	directoryChowned := false
	chown := func(path string, uid, gid int) error {
		if path != targetRoot {
			return nil
		}
		directoryChowned = true
		if uid != target.uid || gid != target.gid {
			return fmt.Errorf("directory owner %d:%d", uid, gid)
		}
		if _, err := os.Lstat(filepath.Join(targetRoot, completeMarker)); err != nil {
			return fmt.Errorf("directory chown before marker rename: %w", err)
		}
		if _, err := os.Lstat(filepath.Join(targetRoot, pendingMarker)); !os.IsNotExist(err) {
			return fmt.Errorf("pending marker still present at directory chown")
		}
		info, err := os.Lstat(targetRoot)
		if err != nil {
			return err
		}
		if info.Mode().Perm() != wantPerm {
			return fmt.Errorf("directory mode %04o at chown, want %04o", info.Mode().Perm(), wantPerm)
		}
		return nil
	}
	if err := projectTarget(targetRoot, source, target, chown); err != nil {
		t.Fatal(err)
	}
	if !directoryChowned {
		t.Fatal("unprivileged directory chown did not run")
	}
	if _, err := os.Lstat(filepath.Join(targetRoot, completeMarker)); err != nil {
		t.Fatalf("complete marker was not published: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(targetRoot, pendingMarker)); !os.IsNotExist(err) {
		t.Fatal("pending marker remains after projection")
	}
	info, err := os.Lstat(targetRoot)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != wantPerm {
		t.Fatalf("directory mode = %04o, want %04o", info.Mode().Perm(), wantPerm)
	}
}

func directoryPermAfterChmod(t *testing.T, mode os.FileMode) os.FileMode {
	t.Helper()
	dir := t.TempDir()
	keepTempDirRemovable(t, dir)
	if err := os.Chmod(dir, mode); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(dir)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if runtime.GOOS != "windows" && perm != 0o500 {
		t.Fatalf("chmod 0500 observed as %04o", perm)
	}
	return perm
}
