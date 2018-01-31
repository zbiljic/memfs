package cmd

import (
	"os"
	"os/signal"
)

// signalTrap traps the registered signals and notifies the caller.
func signalTrap(sig ...os.Signal) <-chan os.Signal {
	// channel to notify the caller.
	trapCh := make(chan os.Signal, 1)

	go func(chan<- os.Signal) {
		// channel to receive signals.
		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)

		// `signal.Notify` registers the given channel to
		// receive notifications of the specified signals.
		signal.Notify(sigCh, sig...)

		// Wait for the signal.
		receivedSignal := <-sigCh

		// Once signal has been received stop signal Notify handler.
		signal.Stop(sigCh)

		// Notify the caller.
		trapCh <- receivedSignal
	}(trapCh)

	return trapCh
}
