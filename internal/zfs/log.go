package zfs

import (
	"errors"
	"io"

	"github.com/go-logr/logr"
)

// writeLogrOut is the type of output the writerLogr will print to.
type writeLogrOut int

const (
	// logStdout outputs to stdout.
	logStdout writeLogrOut = iota
	// logrStderr outputs to stderr.
	logStderr
)

// writerLogr is a io.Writer implementation that writes to a logr.Logger.
type writeLogr struct {
	// log is the logr.Logger to write to.
	logr.Logger

	// std is the type of output the writerLogr will print to.
	std writeLogrOut
}

var _ io.Writer = writeLogr{}

// logWriter create a new writerLogr which implements io.Writer, writing to
// log.Logr. std determines which output the writerLogr will print to.
func logWriter(log logr.Logger, std writeLogrOut) io.Writer { return writeLogr{log, std} }

// Write implements io.Writer.
// Write bytes to the logr.Logger.
func (w writeLogr) Write(b []byte) (int, error) {
	if w.std == logStdout {
		w.Info(string(b))
	} else {
		w.Error(errors.New(string(b)), "")
	}

	return len(b), nil
}
