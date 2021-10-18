//go:build !cgo
// +build !cgo

package flate

import (
	"bytes"
	"io"
	"sync"

	"github.com/klauspost/compress/flate"
)

var writerPool = sync.Pool{
	New: func() interface{} {
		zw, err := flate.NewWriter(io.Discard, flate.BestCompression)
		if err != nil {
			panic(err)
		}
		return zw
	},
}

func Compress(in []byte) ([]byte, error) {
	var b bytes.Buffer

	zw := writerPool.Get().(*flate.Writer)
	defer writerPool.Put(zw)

	zw.Reset(&b)
	zw.Write(in)

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func Decompress(in []byte) ([]byte, error) {
	zr := flate.NewReader(bytes.NewReader(in))
	return io.ReadAll(zr)
}
