package systemtransferartifact

import "bytes"

const maximumAnalysisCaptureBytes = 7 << 20

type analysisCapture struct {
	buffer    bytes.Buffer
	truncated bool
}

func (capture *analysisCapture) Write(value []byte) (int, error) {
	remaining := maximumAnalysisCaptureBytes - capture.buffer.Len()
	if remaining > 0 {
		length := len(value)
		if length > remaining {
			length = remaining
			capture.truncated = true
		}
		_, _ = capture.buffer.Write(value[:length])
	} else if len(value) > 0 {
		capture.truncated = true
	}
	return len(value), nil
}

func (capture *analysisCapture) snapshot() ([]byte, bool) {
	return capture.buffer.Bytes(), capture.truncated
}
