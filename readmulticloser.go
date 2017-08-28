package rwmc

import (
	"github.com/hashicorp/go-multierror"

	"io"
)

// ReadMultiCloser is a wrapper fulfilling io.ReadCloser.
// It's use case is when you need to close multiple Closers when a main Reader is done with reading.
type ReadMultiCloser struct {
	closers []io.Closer
	reader  io.Reader
}

// NewReadMultiCloser creates a new ReadMultiCloser.
// Passed arguments form a queue [reader, closers...].
//
// Usually the code looks similar to:
//
//  r0,_ := os.Open("")
//  r1 := unpack1.Open(r0)
//  r2 := unpack2.Open(r1)
//  r := NewMultiCloser(r2, r1, r0)
//
// Notice, similar thing can be achieved through:
//
//  ioutils.NewWriteCloserWrapper(r, func() error {
//      r2.Close()
//      r1.Close()
//      return r0.Close()
//  })
func NewReadMultiCloser(reader io.ReadCloser, closers ...io.Closer) *ReadMultiCloser {
	p := []io.Closer{reader}
	p = append(p, closers...)
	return &ReadMultiCloser{p, reader}
}

func (r *ReadMultiCloser) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

// Close closes all closers starting from the last pushed.
func (r *ReadMultiCloser) Close() error {
	var errors error
	for _, c := range r.closers {
		if err := c.Close(); err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	return errors
}

func (r *ReadMultiCloser) Push(wc io.ReadCloser) {
	r.reader = wc
	r.closers = append([]io.Closer{wc}, r.closers...)
}
