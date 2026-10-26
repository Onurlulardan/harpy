package systemtransfer

import "io"

type exportBoundedWriter struct {
	destination io.Writer
	maximum     int64
	written     int64
}

func (writer *exportBoundedWriter) Write(value []byte) (int, error) {
	if writer == nil || writer.destination == nil || writer.maximum < 1 ||
		int64(len(value)) > writer.maximum-writer.written {
		return 0, ErrExportLimit
	}
	written, err := writer.destination.Write(value)
	writer.written += int64(written)
	if writer.written > writer.maximum {
		return written, ErrExportLimit
	}
	return written, err
}
