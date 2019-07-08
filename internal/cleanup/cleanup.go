package cleanup

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jeremy-miller/standup-reporter/internal/configuration"
)

func Setup(config *configuration.Configuration) {
	panicRecovery(config)
	handleSignals(config)
}

func panicRecovery(config *configuration.Configuration) {
	if r := recover(); r != nil {
		cleanup(config)
		os.Exit(1)
	}
}

func handleSignals(config *configuration.Configuration) {
	const signalBuffer int = 1
	ch := make(chan os.Signal, signalBuffer)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go signalListener(ch, config)
}

func signalListener(ch chan os.Signal, config *configuration.Configuration) {
	<-ch
	cleanup(config)
	os.Exit(1)
}

func cleanup(config *configuration.Configuration) {
	close(config.Quit)
	config.WG.Wait()
}
