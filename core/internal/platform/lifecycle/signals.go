package lifecycle

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Signals struct {
	channel chan os.Signal
	once    sync.Once
}

func NotifySignals() *Signals {
	channel := make(chan os.Signal, 2)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	return &Signals{channel: channel}
}

func (signals *Signals) Channel() <-chan os.Signal {
	return signals.channel
}

func (signals *Signals) Stop() {
	signals.once.Do(func() {
		signal.Stop(signals.channel)
	})
}
