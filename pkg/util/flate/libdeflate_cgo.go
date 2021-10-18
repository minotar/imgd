//go:build cgo
// +build cgo

package flate

import "github.com/4kills/go-libdeflate"

func Compress(in []byte) ([]byte, error) {
	_, out, err := libdeflate.Compress(in, nil, libdeflate.ModeDEFLATE)
	return out, err
}

func Decompress(in []byte) ([]byte, error) {
	return libdeflate.Decompress(in, nil, libdeflate.ModeDEFLATE)
}
