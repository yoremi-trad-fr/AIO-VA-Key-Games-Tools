package siglus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadRebuildInputFileFallsBackToParent(t *testing.T) {
	root := t.TempDir()
	patched := filepath.Join(root, "patched")
	if err := os.MkdirAll(patched, 0755); err != nil {
		t.Fatal(err)
	}
	want := []byte{0x01, 0x02, 0x03}
	if err := os.WriteFile(filepath.Join(root, "fname.bin"), want, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readRebuildInputFile(patched, "fname.bin")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("fallback data = %v, want %v", got, want)
	}
}

func TestReadRebuildInputFilePrefersInputDir(t *testing.T) {
	root := t.TempDir()
	patched := filepath.Join(root, "patched")
	if err := os.MkdirAll(patched, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "fname.bin"), []byte{0x01}, 0644); err != nil {
		t.Fatal(err)
	}
	want := []byte{0x02}
	if err := os.WriteFile(filepath.Join(patched, "fname.bin"), want, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readRebuildInputFile(patched, "fname.bin")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("input dir data = %v, want %v", got, want)
	}
}

func TestReadRebuildInputFileFallsBackWithRelativeDir(t *testing.T) {
	root := t.TempDir()
	patched := filepath.Join(root, "patched")
	if err := os.MkdirAll(patched, 0755); err != nil {
		t.Fatal(err)
	}
	want := []byte{0x09, 0x08}
	if err := os.WriteFile(filepath.Join(root, "fname.bin"), want, 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	got, err := readRebuildInputFile("patched", "fname.bin")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("relative fallback data = %v, want %v", got, want)
	}
}

func TestResolvePCKWTFAcceptsManualHex(t *testing.T) {
	got, err := ResolvePCKWTF("", "0x166")
	if err != nil {
		t.Fatal(err)
	}
	if got != 0x166 {
		t.Fatalf("wtf = %#x", got)
	}
}

func TestResolvePCKWTFAcceptsUnsignedHex(t *testing.T) {
	got, err := ResolvePCKWTF("", "0xFFFFFFFF")
	if err != nil {
		t.Fatal(err)
	}
	if got != -1 {
		t.Fatalf("wtf = %#x", uint32(got))
	}
}

func TestResolvePCKWTFFromMetadata(t *testing.T) {
	root := t.TempDir()
	meta := PCKMetadata{
		Format: "siglus-pck-metadata-v1",
		WTF:    0x1234,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, pckMetadataFile), data, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ResolvePCKWTF(root, "auto")
	if err != nil {
		t.Fatal(err)
	}
	if got != 0x1234 {
		t.Fatalf("wtf = %#x", got)
	}
}

func TestResolvePCKWTFFallsBackToParentMetadata(t *testing.T) {
	root := t.TempDir()
	patched := filepath.Join(root, "patched")
	if err := os.MkdirAll(patched, 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(PCKMetadata{
		Format: "siglus-pck-metadata-v1",
		WTF:    0x4567,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, pckMetadataFile), data, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ResolvePCKWTF(patched, "metadata")
	if err != nil {
		t.Fatal(err)
	}
	if got != 0x4567 {
		t.Fatalf("wtf = %#x", got)
	}
}
