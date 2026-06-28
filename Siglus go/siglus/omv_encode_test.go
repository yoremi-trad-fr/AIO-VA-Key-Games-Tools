package siglus

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"
)

func TestSequenceLessUsesLeadingNumbers(t *testing.T) {
	names := []string{"10.png", "2.png", "001.png"}
	if !sequenceLess(names[1], names[0]) {
		t.Fatalf("2.png should sort before 10.png")
	}
	if !sequenceLess(names[2], names[1]) {
		t.Fatalf("001.png should sort before 2.png")
	}
}

func TestDefaultPNGVideoOutputAlphaLayout(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 3))
	img.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 40})
	img.SetNRGBA(1, 0, color.NRGBA{R: 11, G: 21, B: 31, A: 41})
	img.SetNRGBA(0, 1, color.NRGBA{R: 12, G: 22, B: 32, A: 42})
	img.SetNRGBA(1, 1, color.NRGBA{R: 13, G: 23, B: 33, A: 43})
	img.SetNRGBA(0, 2, color.NRGBA{R: 14, G: 24, B: 34, A: 44})
	img.SetNRGBA(1, 2, color.NRGBA{R: 15, G: 25, B: 35, A: 45})

	got := imageToSiglusAlphaYUV444(img)
	width := 2
	height := 3
	encodedHeight := alphaEncodedHeight(height)
	if encodedHeight != 4 {
		t.Fatalf("encoded height = %d, want 4", encodedHeight)
	}
	if len(got) != width*encodedHeight*3 {
		t.Fatalf("encoded size = %d", len(got))
	}
	if got[0] != 30 || got[width*encodedHeight] != 20 || got[width*encodedHeight*2] != 10 {
		t.Fatalf("first pixel channels not B/G/R")
	}
	alphaRows := []int{height + 0, height*2 + 1, height*3 + 2}
	wantAlpha := []byte{40, 42, 44}
	for i, row := range alphaRows {
		if got[width*row] != wantAlpha[i] {
			t.Fatalf("alpha row %d = %d, want %d", row, got[width*row], wantAlpha[i])
		}
	}
}

func TestPNGVideoAlphaLayoutWithPadding(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 3))
	img.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 40})
	img.SetNRGBA(0, 1, color.NRGBA{R: 12, G: 22, B: 32, A: 42})
	img.SetNRGBA(0, 2, color.NRGBA{R: 14, G: 24, B: 34, A: 44})

	got := imageToSiglusAlphaYUV444Padded(img, 8)
	width := 2
	if len(got) != width*8*3 {
		t.Fatalf("encoded size = %d", len(got))
	}
	if got[width*(8*0+0)] != 30 || got[width*(8*1+1)] != 22 || got[width*(8*2+2)] != 14 {
		t.Fatalf("padded color planes are not mapped correctly")
	}
	if got[width*3] != 40 || got[width*(8+3)] != 42 || got[width*(16+3)] != 44 {
		t.Fatalf("padded alpha planes are not mapped correctly")
	}
}

func TestDefaultPNGVideoOutputPath(t *testing.T) {
	dir := filepath.Join("C:\\", "frames", "attack")
	got, err := EncodePNGSequenceVideo(dir, "", PNGVideoOptions{})
	if err == nil || got.OutputPath != "" {
		t.Fatalf("expected missing directory error before default output is observable")
	}
}
