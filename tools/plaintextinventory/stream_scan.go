package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

func streamContains(
	ctx context.Context,
	reader io.Reader,
	needle []byte,
	maximum int64,
) (bool, error) {
	if ctx == nil || reader == nil || len(needle) == 0 || maximum < 1 {
		return false, fmt.Errorf("stream scan is invalid")
	}
	buffer := make([]byte, 64*1024+len(needle)-1)
	carry := 0
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		read, err := reader.Read(buffer[carry:])
		total += int64(read)
		if total > maximum {
			return false, fmt.Errorf("stream exceeds scan limit")
		}
		window := buffer[:carry+read]
		if bytes.Contains(window, needle) {
			return true, nil
		}
		if err == io.EOF {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		carry = min(len(needle)-1, len(window))
		copy(buffer[:carry], window[len(window)-carry:])
	}
}
