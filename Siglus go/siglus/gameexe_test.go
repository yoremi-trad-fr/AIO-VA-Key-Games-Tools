package siglus

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGameexeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	key := [16]byte{0x67, 0x2F, 0x7B, 0x9C, 0x82, 0x9B, 0xE9, 0xEB, 0xA9, 0x4F, 0xEF, 0xBF, 0xE4, 0xF8, 0x36, 0xBE}
	ini := append([]byte{0xFF, 0xFE}, []byte{
		'#', 0, 'N', 0, 'A', 0, 'M', 0, 'E', 0, ' ', 0, '=', 0, ' ', 0, '"', 0, 'T', 0, 'e', 0, 's', 0, 't', 0, '"', 0, '\r', 0, '\n', 0,
	}...)
	inPath := filepath.Join(dir, "Gameexe.ini")
	datPath := filepath.Join(dir, "Gameexe.dat")
	outPath := filepath.Join(dir, "Gameexe.out.ini")
	if err := os.WriteFile(inPath, ini, 0644); err != nil {
		t.Fatal(err)
	}
	if err := PackGameexe(inPath, key, true, 17, false, datPath); err != nil {
		t.Fatal(err)
	}
	if err := UnpackGameexe(datPath, key, outPath); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, ini) {
		t.Fatalf("Gameexe round trip mismatch")
	}
}
