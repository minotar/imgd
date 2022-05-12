//go:build cgo
// +build cgo

// When cgo is enabled, we can use the slightly more efficient libdeflate
// Todo: Further optimization would be possible via the libdeflate.NewCompressor

package flate

import "github.com/4kills/go-libdeflate"

func Compress(in []byte) ([]byte, error) {
	_, out, err := libdeflate.Compress(in, nil, libdeflate.ModeDEFLATE)
	if err != nil && err.Error() == "libdeflate: native: compressed data is much larger than uncompressed" {
		// libdeflate complains if the compression made it bigger
		// We "probably" don't care as it was likely a 15 byte invalid user
		return out, nil
	}
	return out, err
}

func Decompress(in []byte) ([]byte, error) {
	return libdeflate.Decompress(in, nil, libdeflate.ModeDEFLATE)
}
