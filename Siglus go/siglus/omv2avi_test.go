package siglus

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertOMV2AVIType2WritesOGV(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "movie.omv")
	ogv := makeTestTheoraOGV()
	omv, err := buildOMV(ogv)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, omv, 0644); err != nil {
		t.Fatal(err)
	}

	defaultOut, err := DefaultOMV2AVIOutput(input)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(defaultOut) != ".ogv" {
		t.Fatalf("default output = %s, want .ogv", defaultOut)
	}

	result, err := ConvertOMV2AVI(input, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.OMVType != 2 || result.OutputPath != defaultOut {
		t.Fatalf("unexpected result: %+v", result)
	}
	got, err := os.ReadFile(defaultOut)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, ogv) {
		t.Fatalf("OGV payload changed")
	}
}

func TestSiglusAlphaYUV444ToBGRA(t *testing.T) {
	const width = 2
	const encodedHeight = 4
	const outputHeight = encodedHeight * 3 / 4
	yuv := make([]byte, width*encodedHeight*3)
	for y := 0; y < encodedHeight; y++ {
		for x := 0; x < width; x++ {
			yuv[width*(encodedHeight*0+y)+x] = byte(10 + y*2 + x)
			yuv[width*(encodedHeight*1+y)+x] = byte(40 + y*2 + x)
			yuv[width*(encodedHeight*2+y)+x] = byte(70 + y*2 + x)
		}
	}

	got, err := siglusAlphaYUV444ToBGRA(yuv, width, encodedHeight)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != width*outputHeight*4 {
		t.Fatalf("BGRA size = %d", len(got))
	}

	wantPixel := func(y, x int) [4]byte {
		third := (outputHeight + 2) / 3
		alphaRow := outputHeight + y
		if y >= third && y < third*2 {
			alphaRow = outputHeight*2 + y
		} else if y >= third*2 {
			alphaRow = outputHeight*3 + y
		}
		return [4]byte{
			yuv[width*(encodedHeight*0+y)+x],
			yuv[width*(encodedHeight*1+y)+x],
			yuv[width*(encodedHeight*2+y)+x],
			yuv[width*alphaRow+x],
		}
	}

	for y := 0; y < outputHeight; y++ {
		for x := 0; x < width; x++ {
			pos := (y*width + x) * 4
			gotPixel := [4]byte{got[pos], got[pos+1], got[pos+2], got[pos+3]}
			if gotPixel != wantPixel(y, x) {
				t.Fatalf("pixel %d,%d = %v, want %v", x, y, gotPixel, wantPixel(y, x))
			}
		}
	}
}
