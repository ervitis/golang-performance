package common

import (
	"os"
	"os/signal"
	"syscall"
)

var SignalHandler = make(chan os.Signal)

func InitSignal() {
	signal.Notify(SignalHandler, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
}
