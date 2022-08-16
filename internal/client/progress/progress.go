package progress

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
)

// out is the writer progress is written to. Used for testing.
var out io.Writer = os.Stdout

// Progress is a progress writer when writing files.
// Implements the io.Reader interface.
type Progress struct {
	// prefix is the prefix of the progress output.
	prefix string

	// r is the input stream to read from.
	r io.Reader

	// size if the given size of file being streamed.
	size uint64

	// read is the number of bytes read so far.
	read uint64
}

// New returns a new Progress, used for reporting the current progress of a
// stream.
func New(prefix string, size uint64, r io.Reader) *Progress {
	return &Progress{prefix: prefix, r: r, size: size}
}

// Read implements io.Reader interface.
func (p *Progress) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if err != nil {
		return -1, err
	}

	atomic.AddUint64(&p.read, uint64(n))
	fmt.Fprintf(out, "%s\t%d/%d (%.2f%%)\n", p.prefix, p.read, p.size, float64(p.read*100)/float64(p.size))

	return n, nil
}
