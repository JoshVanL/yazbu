package signals

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/joshvanl/yazbu/internal/util"
)

var onlyOneSignalHandler = make(chan struct{})

// SetupSignalHandler registers for SIGTERM and SIGINT. A context is returned
// which is canceled on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler(io util.IO) context.Context {
	close(onlyOneSignalHandler) // panics when called twice

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		s := <-c
		fmt.Fprintf(io.Out, "Received signal %s, exiting gracefully...\n", s)
		cancel()
		s = <-c
		fmt.Fprintf(io.Out, "Received second signal %s, exiting ungracefully!\n", s)
		os.Exit(1)
	}()

	return ctx
}
