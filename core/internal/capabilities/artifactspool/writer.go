package artifactspool

import (
	"fmt"
	"os"
)

type boundedWriter struct {
	lease   *Lease
	written int64
	closed  bool
}

func (writer *boundedWriter) Write(value []byte) (int, error) {
	if writer.closed || writer.lease == nil || writer.lease.file == nil {
		return 0, fmt.Errorf("artifact spool writer is closed")
	}
	remaining := writer.lease.manager.reservationBytes - writer.written
	if int64(len(value)) > remaining {
		return 0, fmt.Errorf("artifact spool reservation is exceeded")
	}
	written, err := writer.lease.file.Write(value)
	writer.written += int64(written)
	if err == nil && written != len(value) {
		err = os.ErrInvalid
	}
	return written, err
}

func (writer *boundedWriter) Close() error {
	if writer.closed || writer.lease == nil {
		return nil
	}
	writer.closed = true
	lease := writer.lease
	lease.mu.Lock()
	defer lease.mu.Unlock()
	if lease.file == nil || writer.written < 1 {
		return fmt.Errorf("artifact spool is empty")
	}
	if err := lease.file.Truncate(writer.written); err != nil {
		return fmt.Errorf("truncate artifact spool: %w", err)
	}
	if err := lease.file.Sync(); err != nil {
		return fmt.Errorf("sync artifact spool: %w", err)
	}
	if err := lease.file.Close(); err != nil {
		return fmt.Errorf("close artifact spool: %w", err)
	}
	lease.file = nil
	lease.sealed = true
	return nil
}
