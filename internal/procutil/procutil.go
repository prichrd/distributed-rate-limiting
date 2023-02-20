package procutil

import (
	"os"
	"os/signal"
)

// WaitForInterrupt is a util function that will block until an interrup signal
// is sent to the process.
func WaitForInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// Wait for interrupt
	<-sig

	return
}
