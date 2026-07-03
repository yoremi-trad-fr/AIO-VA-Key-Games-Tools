package siglus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInjectSSDirCopiesRebuildInputs(t *testing.T) {
	root := t.TempDir()
	ssDir := filepath.Join(root, "scene")
	textDir := filepath.Join(root, "text")
	outDir := filepath.Join(root, "patched")
	if err := os.MkdirAll(ssDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(textDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := map[string][]byte{
		"script.ss":  {0x01, 0x02, 0x03},
		"fname.bin":  {0x04, 0x05},
		"table1.bin": {0x06, 0x07},
		"note.txt":   {0x08},
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(ssDir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := InjectSSDir(ssDir, textDir, outDir); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"script.ss", "fname.bin", "table1.bin"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("%s was not copied: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "note.txt")); !os.IsNotExist(err) {
		t.Fatalf("note.txt should not be copied")
	}
}
