package systemtransfer

import (
	"bytes"
	"context"
	"errors"
	"io"
)

type exportSpoolFixture struct {
	buffer       bytes.Buffer
	writerClosed bool
	readerOpened bool
	readerClosed bool
	discarded    bool
}

func (spool *exportSpoolFixture) OpenWriter() (io.WriteCloser, error) {
	if spool.writerClosed || spool.readerOpened || spool.discarded {
		return nil, errors.New("invalid fixture writer state")
	}
	return &exportSpoolWriterFixture{spool: spool}, nil
}

func (spool *exportSpoolFixture) OpenReader() (io.ReadCloser, error) {
	if !spool.writerClosed || spool.readerOpened || spool.discarded {
		return nil, errors.New("invalid fixture reader state")
	}
	spool.readerOpened = true
	return &exportSpoolReaderFixture{Reader: bytes.NewReader(spool.buffer.Bytes()),
		spool: spool}, nil
}

func (spool *exportSpoolFixture) Discard() error {
	if spool.discarded {
		return errors.New("fixture spool discarded twice")
	}
	spool.discarded = true
	spool.buffer.Reset()
	return nil
}

type exportSpoolWriterFixture struct{ spool *exportSpoolFixture }

func (writer *exportSpoolWriterFixture) Write(value []byte) (int, error) {
	if writer.spool.writerClosed {
		return 0, errors.New("fixture writer is closed")
	}
	return writer.spool.buffer.Write(value)
}

func (writer *exportSpoolWriterFixture) Close() error {
	if writer.spool.writerClosed {
		return errors.New("fixture writer closed twice")
	}
	writer.spool.writerClosed = true
	return nil
}

type exportSpoolReaderFixture struct {
	*bytes.Reader
	spool *exportSpoolFixture
}

func (reader *exportSpoolReaderFixture) Close() error {
	if reader.spool.readerClosed {
		return errors.New("fixture reader closed twice")
	}
	reader.spool.readerClosed = true
	return nil
}

type exportSpoolsFixture struct {
	spool    *exportSpoolFixture
	identity ExportPlaintextSpoolIdentity
	maximum  int64
}

func (spools *exportSpoolsFixture) ReserveExportPlaintext(
	_ context.Context,
	identity ExportPlaintextSpoolIdentity,
	maximum int64,
) (ExportPlaintextSpool, error) {
	spools.identity, spools.maximum = identity, maximum
	return spools.spool, nil
}
