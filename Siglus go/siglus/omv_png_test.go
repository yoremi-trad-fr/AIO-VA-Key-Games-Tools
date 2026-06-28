package siglus

import (
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultOMVPNGOutputDirRemovesExtension(t *testing.T) {
	got := DefaultOMVPNGOutputDir(filepath.Join("C:\\", "mov", "attack.omv"))
	want := filepath.Join("C:\\", "mov", "attack")
	if got != want {
		t.Fatalf("output dir = %q, want %q", got, want)
	}
}

func TestWriteBGRAPNGConvertsChannels(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "frame.png")
	bgra := []byte{
		1, 2, 3, 4,
		5, 6, 7, 8,
	}
	if err := writeBGRAPNG(out, bgra, 2, 1); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	p0 := color.NRGBAModel.Convert(img.At(0, 0)).(color.NRGBA)
	if p0.R != 3 || p0.G != 2 || p0.B != 1 || p0.A != 4 {
		t.Fatalf("pixel 0 RGBA = %d,%d,%d,%d", p0.R, p0.G, p0.B, p0.A)
	}
	p1 := color.NRGBAModel.Convert(img.At(1, 0)).(color.NRGBA)
	if p1.R != 7 || p1.G != 6 || p1.B != 5 || p1.A != 8 {
		t.Fatalf("pixel 1 RGBA = %d,%d,%d,%d", p1.R, p1.G, p1.B, p1.A)
	}
}
