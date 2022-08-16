package main

import (
	"fmt"
	"os"

	"github.com/joshvanl/yazbu/internal/cmd"
	"github.com/joshvanl/yazbu/internal/cmd/util/signals"
	"github.com/joshvanl/yazbu/internal/util"
)

func main() {
	io := util.IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr}
	cmd := cmd.New(signals.SetupSignalHandler(io), io)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(io.Err, "%s\n", err)
		os.Exit(1)
	}
}
