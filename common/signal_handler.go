package common

import (
	"os"
	"os/signal"
	"syscall"
)

type Signals struct {
	InterruptSignal chan os.Signal
}

var SignalHandler Signals

func init() {
	SignalHandler = Signals{
		InterruptSignal: make(chan os.Signal, 1),
	}

	signal.Notify(SignalHandler.InterruptSignal, os.Interrupt, os.Kill, syscall.SIGTERM)
}
