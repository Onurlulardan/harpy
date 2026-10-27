//go:build integration

package postgres

import (
	"sync"
	"testing"
)

func runAuthHTTPThrottleRace(t *testing.T, operation func() error) {
	t.Helper()
	start := make(chan struct{})
	errors := make(chan error, 5)
	var group sync.WaitGroup
	for range 5 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			errors <- operation()
		}()
	}
	close(start)
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent throttle failure: %v", err)
		}
	}
}
