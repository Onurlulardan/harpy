package download

import (
	"context"
	"io"
	"sync"
)

type contentStream struct {
	*io.PipeReader
	source io.Closer
	cancel context.CancelFunc
	once   sync.Once
}

func (stream *contentStream) Close() error {
	var result error
	stream.once.Do(func() {
		stream.cancel()
		result = stream.PipeReader.Close()
		if err := stream.source.Close(); result == nil {
			result = err
		}
	})
	return result
}
