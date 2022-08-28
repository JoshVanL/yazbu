package progress

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Progress(t *testing.T) {
	r, w := io.Pipe()

	var buf bytes.Buffer
	out = &buf
	p := New("test", 1000, r)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := w.Write(make([]byte, 100))
			assert.NoError(t, err)
		}()
		_, err := p.Read(make([]byte, 100))
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("test\t%d/1000 (%d.00%%)\n", (i+1)*100, (i+1)*10), buf.String())
		buf.Reset()
	}
}
