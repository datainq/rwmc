package rwmc

import (
	"github.com/hashicorp/go-multierror"

	"io"
)

// WriteMultiCloser is a wrapper fulfilling io.WriteCloser.
// It's use case is when you need to close multiple Closers when a main Writer is done with reading.
type WriteMultiCloser struct {
	closers []io.Closer
	writer  io.Writer
}

// NewWriteMultiCloser creates a new WriteMultiCloser.
// Passed arguments form a queue [reader, closers...].
//
// Usually the code looks similar to:
//
//  w0,_ := os.Create("")
//  w1 := packer.New(w0)
//  w2 := packer.Open(w1)
//  w := NewMultiCloser(w2, w1, r0)
//
func NewWriteMultiCloser(writer io.WriteCloser, closers ...io.Closer) *WriteMultiCloser {
	p := []io.Closer{writer}
	p = append(p, closers...)
	return &WriteMultiCloser{p, writer}
}

func (w *WriteMultiCloser) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *WriteMultiCloser) Close() error {
	var errors error
	for _, c := range w.closers {
		if err := c.Close(); err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	return errors
}

// Push replaces the writer with and adds it to closers.
func (w *WriteMultiCloser) Push(wc io.WriteCloser) {
	w.writer = wc
	w.closers = append([]io.Closer{wc}, w.closers...)
}
