package flate

import (
	"bytes"
	stdflate "compress/flate"
	"io"
	"testing"
)

func TestRawDEFLATECompatibility(t *testing.T) {
	for _, input := range [][]byte{nil, []byte("a short cached value"), bytes.Repeat([]byte("skin"), 10000)} {
		compressed, err := Compress(input)
		if err != nil {
			t.Fatal(err)
		}
		reader := stdflate.NewReader(bytes.NewReader(compressed))
		decoded, err := io.ReadAll(reader)
		reader.Close()
		if err != nil || !bytes.Equal(decoded, input) {
			t.Fatalf("standard decoder: %v", err)
		}

		var buffer bytes.Buffer
		writer, err := stdflate.NewWriter(&buffer, stdflate.BestCompression)
		if err != nil {
			t.Fatal(err)
		}
		writer.Write(input)
		writer.Close()
		decoded, err = Decompress(buffer.Bytes())
		if err != nil || !bytes.Equal(decoded, input) {
			t.Fatalf("decode existing raw DEFLATE: %v", err)
		}
	}
}
