package lifecycle

import "testing"

func TestSignalsUseTwoEventBufferAndStopIdempotently(t *testing.T) {
	signals := NotifySignals()
	if signals.Channel() == nil {
		t.Fatal("signal channel is nil")
	}
	if capacity := cap(signals.channel); capacity != 2 {
		t.Fatalf("signal capacity = %d", capacity)
	}
	signals.Stop()
	signals.Stop()
}
