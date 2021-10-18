//go:build !cgo
// +build !cgo

package flate

import (
	"bytes"
	"compress/flate"
	"io/ioutil"
)

func Compress(in []byte) ([]byte, error) {
	var b bytes.Buffer

	zw, err := flate.NewWriter(&b, flate.BestCompression)
	if err != nil {
		return nil, err
	}

	zw.Write(in)
	if err = zw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func Decompress(in []byte) ([]byte, error) {
	zr := flate.NewReader(bytes.NewReader(in))
	return ioutil.ReadAll(zr)
}
