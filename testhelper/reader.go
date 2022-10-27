package testhelper

import (
	"bytes"
	"io"
)

type StubReader struct {
	ReadN   int
	ReadErr error
}

func (s *StubReader) Read(_ []byte) (n int, err error) {
	return s.ReadN, s.ReadErr
}

var EmptyReadCloser = io.NopCloser(bytes.NewBufferString(""))
