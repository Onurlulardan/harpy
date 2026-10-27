package async

import (
	"testing"
	"time"
)

type fixedJitter struct {
	choose func(uint64) uint64
}

func (source fixedJitter) Uint64N(bound uint64) uint64 {
	return source.choose(bound)
}

func TestRetryDelayUsesInclusiveFullJitterAndCap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		attempt int
		limit   time.Duration
	}{
		{1, 5 * time.Second},
		{2, 10 * time.Second},
		{8, 10*time.Minute + 40*time.Second},
		{9, 15 * time.Minute},
		{25, 15 * time.Minute},
	}
	for _, test := range tests {
		test := test
		t.Run(test.limit.String(), func(t *testing.T) {
			t.Parallel()
			delay, err := RetryDelay(test.attempt, fixedJitter{
				choose: func(bound uint64) uint64 {
					if bound != uint64(test.limit)+1 {
						t.Fatalf("bound = %d, want %d", bound, uint64(test.limit)+1)
					}
					return bound - 1
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if delay != test.limit {
				t.Fatalf("delay = %s, want %s", delay, test.limit)
			}
		})
	}
}

func TestRetryDelayRejectsInvalidInputsAndSource(t *testing.T) {
	t.Parallel()
	invalid := []int{0, 26}
	for _, attempt := range invalid {
		if _, err := RetryDelay(attempt, fixedJitter{}); err == nil {
			t.Fatalf("attempt %d was accepted", attempt)
		}
	}
	if _, err := RetryDelay(1, nil); err == nil {
		t.Fatal("nil jitter source was accepted")
	}
	if _, err := RetryDelay(1, fixedJitter{
		choose: func(bound uint64) uint64 { return bound },
	}); err == nil {
		t.Fatal("out-of-range jitter result was accepted")
	}
}
