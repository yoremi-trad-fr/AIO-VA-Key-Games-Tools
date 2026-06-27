package siglus

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCombinePNGDirOverlaysImagesInNameOrder(t *testing.T) {
	dir := t.TempDir()
	writeTestPNG(t, filepath.Join(dir, "001_base.png"), image.NewUniform(color.RGBA{R: 255, A: 255}), image.Rect(0, 0, 3, 3))

	layer := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	layer.SetNRGBA(0, 0, color.NRGBA{G: 255, A: 255})
	layer.SetNRGBA(1, 0, color.NRGBA{B: 255, A: 128})
	writeTestPNGImage(t, filepath.Join(dir, "002_layer.png"), layer)

	output := filepath.Join(dir, "result.png")
	if err := CombinePNGDir(dir, output); err != nil {
		t.Fatal(err)
	}
	got := readTestPNG(t, output)

	if rgba := color.RGBAModel.Convert(got.At(0, 0)).(color.RGBA); rgba != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("opaque overlay pixel = %#v", rgba)
	}
	blended := color.RGBAModel.Convert(got.At(1, 0)).(color.RGBA)
	if blended.R < 126 || blended.R > 128 || blended.B < 127 || blended.B > 129 || blended.A != 255 {
		t.Fatalf("semi-transparent overlay pixel = %#v", blended)
	}
	if rgba := color.RGBAModel.Convert(got.At(2, 2)).(color.RGBA); rgba != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("base-only pixel = %#v", rgba)
	}
}

func TestDefaultCombinePNGOutputUsesDirectoryName(t *testing.T) {
	got := DefaultCombinePNGOutput(filepath.Join("C:\\", "work", "layers"))
	want := filepath.Join("C:\\", "work", "layers", "layers.png")
	if got != want {
		t.Fatalf("default output = %q, want %q", got, want)
	}
}

func writeTestPNG(t *testing.T, path string, src image.Image, bounds image.Rectangle) {
	t.Helper()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, src.At(x, y))
		}
	}
	writeTestPNGImage(t, path, img)
}

func writeTestPNGImage(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func readTestPNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}
