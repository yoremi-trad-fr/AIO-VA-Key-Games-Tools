package siglus

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DefaultCombinePNGOutput(inputDir string) string {
	clean := filepath.Clean(inputDir)
	return filepath.Join(clean, filepath.Base(clean)+".png")
}

func CombinePNGDir(inputDir, outputPath string) error {
	if strings.TrimSpace(outputPath) == "" {
		outputPath = DefaultCombinePNGOutput(inputDir)
	}
	files, err := listCombinePNGs(inputDir, outputPath)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no PNG files found in %s", inputDir)
	}

	base, err := decodePNG(files[0])
	if err != nil {
		return err
	}
	canvas := image.NewRGBA(base.Bounds())
	draw.Draw(canvas, canvas.Bounds(), base, base.Bounds().Min, draw.Src)

	for _, file := range files[1:] {
		layer, err := decodePNG(file)
		if err != nil {
			return err
		}
		draw.Draw(canvas, layer.Bounds().Add(base.Bounds().Min), layer, layer.Bounds().Min, draw.Over)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output PNG: %w", err)
	}
	defer out.Close()
	if err := png.Encode(out, canvas); err != nil {
		return fmt.Errorf("cannot encode output PNG: %w", err)
	}
	return nil
}

func listCombinePNGs(inputDir, outputPath string) ([]string, error) {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read PNG directory: %w", err)
	}
	outputAbs, _ := filepath.Abs(outputPath)
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			continue
		}
		path := filepath.Join(inputDir, entry.Name())
		abs, _ := filepath.Abs(path)
		if outputAbs != "" && strings.EqualFold(abs, outputAbs) {
			continue
		}
		files = append(files, path)
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(filepath.Base(files[i])) < strings.ToLower(filepath.Base(files[j]))
	})
	return files, nil
}

func decodePNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open PNG %s: %w", path, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("cannot decode PNG %s: %w", path, err)
	}
	return img, nil
}
