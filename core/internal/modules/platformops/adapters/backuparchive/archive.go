package backuparchive

import (
	"context"
	"errors"
	"io"
)

var ErrArchive = errors.New("database recovery archive operation failed")

type Dumper interface {
	Dump(context.Context, string, string, io.Writer) error
}

type Restorer interface {
	Restore(context.Context, string, io.Reader) error
}

type Adapter struct {
	dumper   Dumper
	restorer Restorer
}

func New(dumper Dumper, restorer Restorer) *Adapter {
	if dumper == nil || restorer == nil {
		return nil
	}
	return &Adapter{dumper: dumper, restorer: restorer}
}
