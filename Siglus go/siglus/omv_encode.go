package siglus

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type PNGVideoOptions struct {
	Alpha      bool
	FPS        string
	FFmpegPath string
}

type PNGVideoResult struct {
	OutputPath string
	FrameCount int
	Width      int
	Height     int
	Alpha      bool
}

func EncodePNGSequenceVideo(inputDir, outputPath string, opts PNGVideoOptions) (PNGVideoResult, error) {
	files, err := listSequencePNGs(inputDir)
	if err != nil {
		return PNGVideoResult{}, err
	}
	if len(files) == 0 {
		return PNGVideoResult{}, fmt.Errorf("no PNG files found in %s", inputDir)
	}
	if strings.TrimSpace(outputPath) == "" {
		outputPath = filepath.Join(inputDir, filepath.Base(filepath.Clean(inputDir))+".omv")
	}
	if opts.FPS == "" {
		opts.FPS = "30"
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return PNGVideoResult{}, fmt.Errorf("cannot create output directory: %w", err)
	}

	first, err := decodeSequencePNG(files[0])
	if err != nil {
		return PNGVideoResult{}, err
	}
	width := first.Bounds().Dx()
	height := first.Bounds().Dy()
	if width <= 0 || height <= 0 {
		return PNGVideoResult{}, fmt.Errorf("invalid PNG dimensions")
	}

	tmpDir, err := os.MkdirTemp("", "siglus-png-video-")
	if err != nil {
		return PNGVideoResult{}, fmt.Errorf("cannot create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ogvPath := outputPath
	if strings.EqualFold(filepath.Ext(outputPath), ".omv") {
		ogvPath = filepath.Join(tmpDir, "encoded.ogv")
	}
	if opts.Alpha {
		err = encodeAlphaPNGSequenceToOGV(files, ogvPath, width, height, opts)
	} else {
		err = encodeNormalPNGSequenceToOGV(files, ogvPath, width, height, opts)
	}
	if err != nil {
		return PNGVideoResult{}, err
	}

	if strings.EqualFold(filepath.Ext(outputPath), ".omv") {
		ogv, err := os.ReadFile(ogvPath)
		if err != nil {
			return PNGVideoResult{}, fmt.Errorf("cannot read encoded OGV: %w", err)
		}
		buildOpts := OMVBuildOptions{EncryptType: 2}
		if opts.Alpha {
			buildOpts = OMVBuildOptions{
				EncryptType: 1,
				Width:       uint32(width),
				Height:      uint32(height),
			}
		}
		omv, err := buildOMVWithOptions(ogv, buildOpts)
		if err != nil {
			return PNGVideoResult{}, err
		}
		if err := os.WriteFile(outputPath, omv, 0644); err != nil {
			return PNGVideoResult{}, fmt.Errorf("cannot write OMV: %w", err)
		}
	}

	return PNGVideoResult{
		OutputPath: outputPath,
		FrameCount: len(files),
		Width:      width,
		Height:     height,
		Alpha:      opts.Alpha,
	}, nil
}

func encodeNormalPNGSequenceToOGV(files []string, outputPath string, width, height int, opts PNGVideoOptions) error {
	ffmpeg, err := resolveFFmpeg(opts.FFmpegPath)
	if err != nil {
		return err
	}
	cmd := exec.Command(ffmpeg,
		"-y",
		"-v", "error",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-s:v", fmt.Sprintf("%dx%d", width, height),
		"-r", opts.FPS,
		"-i", "-",
		"-an",
		"-c:v", "libtheora",
		"-pix_fmt", "yuv444p",
		"-g", "1",
		"-q:v", "10",
		outputPath,
	)
	return feedPNGFramesToFFmpeg(cmd, files, width, height, false, height)
}

func encodeAlphaPNGSequenceToOGV(files []string, outputPath string, width, height int, opts PNGVideoOptions) error {
	ffmpeg, err := resolveFFmpeg(opts.FFmpegPath)
	if err != nil {
		return err
	}
	encodedHeight := alphaEncodedHeight(height)
	codedHeight := alignTo(encodedHeight, 16)
	cmd := exec.Command(ffmpeg,
		"-y",
		"-v", "error",
		"-f", "rawvideo",
		"-pix_fmt", "yuv444p",
		"-s:v", fmt.Sprintf("%dx%d", width, codedHeight),
		"-r", opts.FPS,
		"-i", "-",
		"-vf", fmt.Sprintf("crop=%d:%d:0:0", width, encodedHeight),
		"-an",
		"-c:v", "libtheora",
		"-pix_fmt", "yuv444p",
		"-g", "1",
		"-q:v", "10",
		outputPath,
	)
	return feedPNGFramesToFFmpeg(cmd, files, width, height, true, codedHeight)
}

func feedPNGFramesToFFmpeg(cmd *exec.Cmd, files []string, width, height int, alpha bool, codedHeight int) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdin: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stderr: %w", err)
	}
	capturedStderr := capturePipe(stderr)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot start ffmpeg encoder: %w", err)
	}

	for _, file := range files {
		img, err := decodeSequencePNG(file)
		if err != nil {
			stdin.Close()
			cmd.Wait()
			return err
		}
		if img.Bounds().Dx() != width || img.Bounds().Dy() != height {
			stdin.Close()
			cmd.Wait()
			return fmt.Errorf("PNG dimensions differ in %s", file)
		}
		var frame []byte
		if alpha {
			frame = imageToSiglusAlphaYUV444Padded(img, codedHeight)
		} else {
			frame = imageToRGBA(img)
		}
		if _, err := stdin.Write(frame); err != nil {
			stdin.Close()
			cmd.Wait()
			return fmt.Errorf("cannot feed ffmpeg encoder: %w%s", err, formatProcessStderr(<-capturedStderr))
		}
	}
	if err := stdin.Close(); err != nil {
		cmd.Wait()
		return fmt.Errorf("cannot close ffmpeg encoder input: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg encoder failed: %w%s", err, formatProcessStderr(<-capturedStderr))
	}
	return nil
}

func listSequencePNGs(inputDir string) ([]string, error) {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read PNG directory: %w", err)
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			continue
		}
		files = append(files, filepath.Join(inputDir, entry.Name()))
	}
	sort.Slice(files, func(i, j int) bool {
		return sequenceLess(filepath.Base(files[i]), filepath.Base(files[j]))
	})
	return files, nil
}

func sequenceLess(a, b string) bool {
	ai, aok := leadingNumber(a)
	bi, bok := leadingNumber(b)
	if aok && bok && ai != bi {
		return ai < bi
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

func leadingNumber(name string) (int, bool) {
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	if stem == "" {
		return 0, false
	}
	i := 0
	for i < len(stem) && stem[i] >= '0' && stem[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, false
	}
	v, err := strconv.Atoi(stem[:i])
	return v, err == nil
}

func decodeSequencePNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open PNG %s: %w", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("cannot decode PNG %s: %w", path, err)
	}
	return img, nil
}

func imageToRGBA(img image.Image) []byte {
	b := img.Bounds()
	out := make([]byte, b.Dx()*b.Dy()*4)
	pos := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			p := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			out[pos+0] = p.R
			out[pos+1] = p.G
			out[pos+2] = p.B
			out[pos+3] = p.A
			pos += 4
		}
	}
	return out
}

func imageToSiglusAlphaYUV444(img image.Image) []byte {
	b := img.Bounds()
	height := b.Dy()
	return imageToSiglusAlphaYUV444Padded(img, alphaEncodedHeight(height))
}

func imageToSiglusAlphaYUV444Padded(img image.Image, planeHeight int) []byte {
	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()
	encodedHeight := alphaEncodedHeight(height)
	if planeHeight < encodedHeight {
		planeHeight = encodedHeight
	}
	out := make([]byte, width*planeHeight*3)
	third := (height + 2) / 3

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			out[width*(planeHeight*0+y)+x] = p.B
			out[width*(planeHeight*1+y)+x] = p.G
			out[width*(planeHeight*2+y)+x] = p.R

			alphaRow := height + y
			if y >= third && y < third*2 {
				alphaRow = height*2 + y
			} else if y >= third*2 {
				alphaRow = height*3 + y
			}
			alphaPlane := alphaRow / encodedHeight
			alphaY := alphaRow % encodedHeight
			out[width*(planeHeight*alphaPlane+alphaY)+x] = p.A
		}
	}
	return out
}

func alphaEncodedHeight(height int) int {
	return height + (height+2)/3
}

func alignTo(value, align int) int {
	if align <= 0 || value%align == 0 {
		return value
	}
	return value + align - value%align
}
