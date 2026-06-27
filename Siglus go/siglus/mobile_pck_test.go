package siglus

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMobilePCKRoundTrip(t *testing.T) {
	dir := t.TempDir()
	inDir := filepath.Join(dir, "in")
	outDir := filepath.Join(dir, "out")
	pckPath := filepath.Join(dir, "archive.pck")
	if err := os.MkdirAll(inDir, 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{
		"alpha.txt": []byte("alpha"),
		"日本語.bin":   {0x00, 0x01, 0xFE, 0xFF},
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(inDir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := PackMobilePCK(inDir, pckPath); err != nil {
		t.Fatal(err)
	}
	if err := UnpackMobilePCK(pckPath, outDir); err != nil {
		t.Fatal(err)
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s mismatch", name)
		}
	}
}

func TestSafeArchivePathRejectsTraversal(t *testing.T) {
	if _, err := safeArchivePath(t.TempDir(), "..\\escape.txt"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}
