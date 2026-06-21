package siglus

import (
	"bytes"
	"testing"
)

func TestCompressDecompressRoundTrip(t *testing.T) {
	samples := [][]byte{
		[]byte(""),
		[]byte("abcdef"),
		[]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		[]byte("SiglusEngine compression round trip test. SiglusEngine compression round trip test."),
		bytes.Repeat([]byte{0x00, 0x01, 0x02, 0x03, 0x02, 0x01}, 128),
		bytes.Repeat([]byte("0123456789ABCDEF repeated payload "), 512),
	}

	for _, sample := range samples {
		compressed := Compress(sample)
		decompressed := Decompress(compressed)
		if !bytes.Equal(decompressed, sample) {
			t.Fatalf("round trip mismatch for %d-byte sample", len(sample))
		}
	}
}
