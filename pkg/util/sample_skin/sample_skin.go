package sample_skin

//go:generate go run load_skin.go

import (
	"bytes"
	"encoding/base64"
	"io"
)

func GetSampleSkinReadCloser() (io.ReadCloser, error) {
	skinBytes, err := base64.StdEncoding.DecodeString(SampleSkinBase64)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewBuffer(skinBytes)), nil
}
