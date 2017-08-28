package rwmc

import (
	"io"
	"testing"

	"bytes"
	"io/ioutil"

	"github.com/docker/docker/pkg/ioutils"
)

type fakeRC struct {
	closed int
	read   int
	reader io.Reader
}

func (f *fakeRC) Read(p []byte) (n int, err error) {
	f.read++
	return f.reader.Read(p)
}

func (f *fakeRC) Close() error {
	f.closed++
	return nil
}

func TestReadMultiCloser_Read(t *testing.T) {
	exp := []byte("napis")
	{
		r := bytes.NewReader(exp)
		rwc := NewReadMultiCloser(ioutil.NopCloser(r))
		b, err := ioutil.ReadAll(rwc)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !bytes.Equal(exp, b) {
			t.Errorf("want: %x, got: %x", exp, b)
		}
	}
	{
		r := bytes.NewReader(exp)
		rwc := &ReadMultiCloser{}
		rwc.Push(ioutil.NopCloser(r))
		b, err := ioutil.ReadAll(rwc)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !bytes.Equal(exp, b) {
			t.Errorf("want: %x, got: %x", exp, b)
		}
	}
}

func TestReadMultiCloser_Push(t *testing.T) {
	closed0 := false
	rc0 := ioutils.NewReadCloserWrapper(nil, func() error {
		closed0 = true
		return io.ErrUnexpectedEOF
	})

	closed1 := false
	rc1 := ioutils.NewReadCloserWrapper(nil, func() error {
		closed1 = true
		return io.EOF
	})
	rwc := &ReadMultiCloser{}
	rwc.Push(rc1)
	rwc.Push(rc0)

	if err := rwc.Close(); err == nil {
		t.Errorf("want: %v, got: %v", io.ErrUnexpectedEOF, err)
	}
	if !closed0 {
		t.Error("Close was not called on ReadCloser")
	}
	if !closed1 {
		t.Error("Close was not called on Closer")
	}
}

func TestReadMultiCloser_CloseOne(t *testing.T) {
	closed := false
	rc := ioutils.NewReadCloserWrapper(nil, func() error {
		closed = true
		return io.ErrUnexpectedEOF
	})
	rwc := NewReadMultiCloser(rc)
	if err := rwc.Close(); err == nil {
		t.Errorf("want: %v, got: %v", io.ErrUnexpectedEOF, err)
	}
	if !closed {
		t.Error("Close was not called")
	}
}

func TestReadMultiCloser_CloseMany(t *testing.T) {
	closed0 := false
	rc := ioutils.NewReadCloserWrapper(nil, func() error {
		closed0 = true
		return io.ErrUnexpectedEOF
	})

	closed1 := false
	c := ioutils.NewReadCloserWrapper(nil, func() error {
		closed1 = true
		return io.EOF
	})
	rwc := NewReadMultiCloser(rc, c)
	if err := rwc.Close(); err == nil {
		t.Errorf("want: %v, got: %v", io.ErrUnexpectedEOF, err)
	}
	if !closed0 {
		t.Error("Close was not called on ReadCloser")
	}
	if !closed1 {
		t.Error("Close was not called on Closer")
	}
}
