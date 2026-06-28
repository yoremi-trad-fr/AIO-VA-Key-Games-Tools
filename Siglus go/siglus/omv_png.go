package siglus

import (
	"bufio"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type OMVPNGResult struct {
	OutputDir  string
	OMVType    uint32
	FrameCount int
	Width      int
	Height     int
}

func DefaultOMVPNGOutputDir(inputPath string) string {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return inputPath
	}
	return strings.TrimSuffix(inputPath, ext)
}

func ExtractOMVToPNGSequence(inputPath, outputDir, ffmpegPath string) (OMVPNGResult, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return OMVPNGResult{}, fmt.Errorf("cannot read OMV: %w", err)
	}
	info, ogg, err := parseOMVPayload(data)
	if err != nil {
		return OMVPNGResult{}, err
	}
	if strings.TrimSpace(outputDir) == "" {
		outputDir = DefaultOMVPNGOutputDir(inputPath)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return OMVPNGResult{}, fmt.Errorf("cannot create PNG output directory: %w", err)
	}

	switch info.encryptType {
	case 1:
		frames, width, height, err := extractAlphaOMVPNG(ogg, outputDir, ffmpegPath)
		if err != nil {
			return OMVPNGResult{}, err
		}
		return OMVPNGResult{
			OutputDir:  outputDir,
			OMVType:    info.encryptType,
			FrameCount: frames,
			Width:      width,
			Height:     height,
		}, nil
	case 2:
		frames, err := extractNormalOMVPNG(ogg, outputDir, ffmpegPath)
		if err != nil {
			return OMVPNGResult{}, err
		}
		return OMVPNGResult{
			OutputDir:  outputDir,
			OMVType:    info.encryptType,
			FrameCount: frames,
			Width:      int(info.width),
			Height:     int(info.height),
		}, nil
	default:
		return OMVPNGResult{}, fmt.Errorf("unsupported OMV type %d", info.encryptType)
	}
}

func extractAlphaOMVPNG(ogg []byte, outputDir, ffmpegPath string) (int, int, int, error) {
	ffmpeg, err := resolveFFmpeg(ffmpegPath)
	if err != nil {
		return 0, 0, 0, err
	}
	tmpDir, err := os.MkdirTemp("", "siglus-omv-png-")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("cannot create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ogvPath := filepath.Join(tmpDir, "input.ogv")
	if err := os.WriteFile(ogvPath, ogg, 0644); err != nil {
		return 0, 0, 0, fmt.Errorf("cannot write temporary OGV: %w", err)
	}

	cmd := exec.Command(ffmpeg, "-v", "error", "-i", ogvPath, "-f", "yuv4mpegpipe", "-pix_fmt", "yuv444p", "-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("ffmpeg stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("ffmpeg stderr: %w", err)
	}
	capturedStderr := capturePipe(stderr)
	if err := cmd.Start(); err != nil {
		return 0, 0, 0, fmt.Errorf("cannot start ffmpeg decoder: %w", err)
	}

	reader := bufio.NewReaderSize(stdout, 1024*1024)
	yinfo, err := readY4MHeader(reader)
	if err != nil {
		cmd.Wait()
		return 0, 0, 0, fmt.Errorf("cannot read decoded video header: %w%s", err, formatProcessStderr(<-capturedStderr))
	}
	if yinfo.chroma != "" && yinfo.chroma != "444" && !strings.HasPrefix(yinfo.chroma, "444") {
		cmd.Wait()
		return 0, 0, 0, fmt.Errorf("unsupported decoded pixel type %q", yinfo.chroma)
	}

	frameSize := yinfo.width * yinfo.height * 3
	yuv := make([]byte, frameSize)
	frameCount := 0
	for {
		ok, err := readY4MFrame(reader, yuv)
		if err != nil {
			cmd.Wait()
			return 0, 0, 0, err
		}
		if !ok {
			break
		}
		bgra, err := siglusAlphaYUV444ToBGRA(yuv, yinfo.width, yinfo.height)
		if err != nil {
			cmd.Wait()
			return 0, 0, 0, err
		}
		outputHeight := yinfo.height * 3 / 4
		if err := writeBGRAPNG(filepath.Join(outputDir, fmt.Sprintf("%06d.png", frameCount)), bgra, yinfo.width, outputHeight); err != nil {
			cmd.Wait()
			return 0, 0, 0, err
		}
		frameCount++
	}
	if err := cmd.Wait(); err != nil {
		return 0, 0, 0, fmt.Errorf("ffmpeg decoder failed: %w%s", err, formatProcessStderr(<-capturedStderr))
	}
	return frameCount, yinfo.width, yinfo.height * 3 / 4, nil
}

func extractNormalOMVPNG(ogg []byte, outputDir, ffmpegPath string) (int, error) {
	ffmpeg, err := resolveFFmpeg(ffmpegPath)
	if err != nil {
		return 0, err
	}
	tmpDir, err := os.MkdirTemp("", "siglus-omv-png-")
	if err != nil {
		return 0, fmt.Errorf("cannot create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ogvPath := filepath.Join(tmpDir, "input.ogv")
	if err := os.WriteFile(ogvPath, ogg, 0644); err != nil {
		return 0, fmt.Errorf("cannot write temporary OGV: %w", err)
	}

	pattern := filepath.Join(outputDir, "%06d.png")
	cmd := exec.Command(ffmpeg, "-y", "-v", "error", "-i", ogvPath, "-start_number", "0", pattern)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg stderr: %w", err)
	}
	capturedStderr := capturePipe(stderr)
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("cannot start ffmpeg decoder: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return 0, fmt.Errorf("ffmpeg PNG extraction failed: %w%s", err, formatProcessStderr(<-capturedStderr))
	}
	return countPNGFiles(outputDir), nil
}

func writeBGRAPNG(path string, bgra []byte, width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid PNG dimensions")
	}
	if len(bgra) < width*height*4 {
		return fmt.Errorf("BGRA frame is too small")
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			src := (y*width + x) * 4
			dst := y*img.Stride + x*4
			img.Pix[dst+0] = bgra[src+2]
			img.Pix[dst+1] = bgra[src+1]
			img.Pix[dst+2] = bgra[src+0]
			img.Pix[dst+3] = bgra[src+3]
		}
	}
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create PNG: %w", err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		return fmt.Errorf("cannot encode PNG: %w", err)
	}
	return nil
}

func countPNGFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".png") {
			count++
		}
	}
	return count
}
