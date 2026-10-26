package passwordblocklist

import (
	"context"
	"errors"
	"sync"
	"testing"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func TestConcurrentMembershipAndCloseRemainFailClosed(t *testing.T) {
	path, digest := blocklistFixture(t, blocklistv1.KindCommon, "known password")
	blocklist, err := Open(path, digest, blocklistv1.KindCommon)
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	errorsSeen := make(chan error, 16)
	var wait sync.WaitGroup
	for worker := 0; worker < cap(errorsSeen); worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			for attempt := 0; attempt < 100; attempt++ {
				_, lookupErr := blocklist.Blocked(
					context.Background(), []byte("known password"),
				)
				if lookupErr != nil && !errors.Is(lookupErr, ErrUnavailable) {
					errorsSeen <- lookupErr
					return
				}
			}
		}()
	}
	close(start)
	if err := blocklist.Close(); err != nil {
		t.Fatal(err)
	}
	wait.Wait()
	close(errorsSeen)
	for lookupErr := range errorsSeen {
		t.Fatalf("unexpected concurrent lookup error: %v", lookupErr)
	}
}
