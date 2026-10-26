package pgtools

import (
	"context"
	"io"
)

type Command struct {
	Executable  string
	Arguments   []string
	Environment []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

type Runner interface {
	Run(context.Context, Command) error
}

type ArchiveTools struct {
	runner    Runner
	pgDump    string
	pgRestore string
}

func New(runner Runner, pgDump, pgRestore string) *ArchiveTools {
	if runner == nil || !validExecutable(pgDump) || !validExecutable(pgRestore) {
		return nil
	}
	return &ArchiveTools{runner: runner, pgDump: pgDump, pgRestore: pgRestore}
}

func validExecutable(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}
