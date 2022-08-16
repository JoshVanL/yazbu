package util

import "io"

// IO is a utility struct for io.Reader and io.Writer.
type IO struct {
	// In is the input stream.
	In io.Reader

	// Out is the standard output stream.
	Out io.Writer

	// Err is the standard error stream.
	Err io.Writer
}
