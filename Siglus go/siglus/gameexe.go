package siglus

import (
	"encoding/binary"
	"fmt"
	"os"
)

// UnpackGameexe decrypts and decompresses a Siglus Gameexe.dat into UTF-16 LE INI.
func UnpackGameexe(inputPath string, key [16]byte, outputPath string) error {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read Gameexe.dat: %w", err)
	}
	if len(buf) < 16 {
		return fmt.Errorf("file too small to be Gameexe.dat")
	}

	needsKey := binary.LittleEndian.Uint32(buf[4:8]) == 1
	data := append([]byte(nil), buf[8:]...)
	xorGameexe(data)
	if needsKey {
		xorKey16(data, key)
	}

	decompressed := Decompress(data)
	if decompressed == nil {
		return fmt.Errorf("Gameexe.dat decompression failed")
	}

	out := make([]byte, 0, len(decompressed)+2)
	out = append(out, 0xFF, 0xFE)
	out = append(out, decompressed...)
	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		return fmt.Errorf("cannot write Gameexe.ini: %w", err)
	}
	return nil
}

// PackGameexe compresses and encrypts a UTF-16 LE INI into Siglus Gameexe.dat.
func PackGameexe(inputPath string, key [16]byte, useKey bool, compressionLevel int, fakeCompression bool, outputPath string) error {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read Gameexe.ini: %w", err)
	}
	if len(buf) >= 2 && buf[0] == 0xFF && buf[1] == 0xFE {
		buf = buf[2:]
	}

	var data []byte
	if fakeCompression {
		data = FakeCompress(buf)
	} else {
		data = CompressLevel(buf, compressionLevel)
	}
	if useKey {
		xorKey16(data, key)
	}
	xorGameexe(data)

	out := make([]byte, 8, len(data)+8)
	if useKey {
		binary.LittleEndian.PutUint32(out[4:8], 1)
	}
	out = append(out, data...)
	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		return fmt.Errorf("cannot write Gameexe.dat: %w", err)
	}
	return nil
}
