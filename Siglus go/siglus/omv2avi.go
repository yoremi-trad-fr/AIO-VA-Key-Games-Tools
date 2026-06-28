package siglus

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type OMVConvertResult struct {
	OutputPath string
	OMVType    uint32
	FrameCount int
	UsedFFmpeg bool
	Width      int
	Height     int
	EncodedW   int
	EncodedH   int
	FrameRate  string
}

type omvHeaderInfo struct {
	dataOffset       uint32
	mainVersion      byte
	subVersion       byte
	encryptType      uint32
	width            uint32
	height           uint32
	frameTime        uint32
	dataPackageCount uint32
	frameCount       uint32
	oggOffset        int
}

type y4mInfo struct {
	width     int
	height    int
	frameRate string
	chroma    string
}

func DefaultOMV2AVIOutput(inputPath string) (string, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("cannot read OMV: %w", err)
	}
	info, _, err := parseOMVPayload(data)
	if err != nil {
		return "", err
	}
	ext := ".avi"
	if info.encryptType == 2 {
		ext = ".ogv"
	}
	return replaceExtension(inputPath, ext), nil
}

func ConvertOMV2AVI(inputPath, outputPath, ffmpegPath string) (OMVConvertResult, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return OMVConvertResult{}, fmt.Errorf("cannot read OMV: %w", err)
	}
	info, ogg, err := parseOMVPayload(data)
	if err != nil {
		return OMVConvertResult{}, err
	}
	if strings.TrimSpace(outputPath) == "" {
		outputPath = replaceExtension(inputPath, ".avi")
		if info.encryptType == 2 {
			outputPath = replaceExtension(inputPath, ".ogv")
		}
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return OMVConvertResult{}, fmt.Errorf("cannot create output directory: %w", err)
	}

	result := OMVConvertResult{
		OutputPath: outputPath,
		OMVType:    info.encryptType,
		Width:      int(info.width),
		Height:     int(info.height),
	}

	switch info.encryptType {
	case 2:
		if err := os.WriteFile(outputPath, ogg, 0644); err != nil {
			return OMVConvertResult{}, fmt.Errorf("cannot write OGV: %w", err)
		}
		return result, nil
	case 1:
		frames, yinfo, err := convertOMVType1ToAVI(ogg, outputPath, ffmpegPath)
		if err != nil {
			return OMVConvertResult{}, err
		}
		result.FrameCount = frames
		result.UsedFFmpeg = true
		result.EncodedW = yinfo.width
		result.EncodedH = yinfo.height
		result.FrameRate = yinfo.frameRate
		if yinfo.width > 0 {
			result.Width = yinfo.width
		}
		if yinfo.height > 0 {
			result.Height = yinfo.height * 3 / 4
		}
		return result, nil
	default:
		return OMVConvertResult{}, fmt.Errorf("unsupported OMV type %d", info.encryptType)
	}
}

func parseOMVPayload(data []byte) (omvHeaderInfo, []byte, error) {
	if len(data) < 0x58 {
		return omvHeaderInfo{}, nil, fmt.Errorf("file too small to be a valid OMV")
	}
	info := omvHeaderInfo{
		dataOffset:       binary.LittleEndian.Uint32(data[0x00:0x04]),
		mainVersion:      data[0x04],
		subVersion:       data[0x05],
		encryptType:      binary.LittleEndian.Uint32(data[0x28:0x2c]),
		width:            binary.LittleEndian.Uint32(data[0x2c:0x30]),
		height:           binary.LittleEndian.Uint32(data[0x30:0x34]),
		frameTime:        binary.LittleEndian.Uint32(data[0x3c:0x40]),
		dataPackageCount: binary.LittleEndian.Uint32(data[0x4c:0x50]),
		frameCount:       binary.LittleEndian.Uint32(data[0x50:0x54]),
	}
	if info.mainVersion != 1 {
		return omvHeaderInfo{}, nil, fmt.Errorf("unsupported OMV main version %d", info.mainVersion)
	}
	offset := int(info.dataOffset)
	switch info.subVersion {
	case 0:
	case 1:
		offset += int(info.dataPackageCount)*28 + int(info.frameCount)*32
	default:
		return omvHeaderInfo{}, nil, fmt.Errorf("unsupported OMV sub version %d", info.subVersion)
	}
	if offset < 0 || offset+4 > len(data) {
		return omvHeaderInfo{}, nil, fmt.Errorf("OMV data offset out of bounds")
	}
	if !bytes.Equal(data[offset:offset+4], []byte("OggS")) {
		return omvHeaderInfo{}, nil, fmt.Errorf("not an OMV Ogg stream at 0x%x", offset)
	}
	info.oggOffset = offset
	return info, data[offset:], nil
}

func convertOMVType1ToAVI(ogg []byte, outputPath, ffmpegPath string) (int, y4mInfo, error) {
	ffmpeg, err := resolveFFmpeg(ffmpegPath)
	if err != nil {
		return 0, y4mInfo{}, err
	}

	tmpDir, err := os.MkdirTemp("", "siglus-omv2avi-")
	if err != nil {
		return 0, y4mInfo{}, fmt.Errorf("cannot create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ogvPath := filepath.Join(tmpDir, "input.ogv")
	if err := os.WriteFile(ogvPath, ogg, 0644); err != nil {
		return 0, y4mInfo{}, fmt.Errorf("cannot write temporary OGV: %w", err)
	}

	decode := exec.Command(ffmpeg, "-v", "error", "-i", ogvPath, "-f", "yuv4mpegpipe", "-pix_fmt", "yuv444p", "-")
	decodeOut, err := decode.StdoutPipe()
	if err != nil {
		return 0, y4mInfo{}, fmt.Errorf("ffmpeg decode stdout: %w", err)
	}
	decodeErr, err := decode.StderrPipe()
	if err != nil {
		return 0, y4mInfo{}, fmt.Errorf("ffmpeg decode stderr: %w", err)
	}
	decodeStderr := capturePipe(decodeErr)
	if err := decode.Start(); err != nil {
		return 0, y4mInfo{}, fmt.Errorf("cannot start ffmpeg decoder: %w", err)
	}

	reader := bufio.NewReaderSize(decodeOut, 1024*1024)
	yinfo, err := readY4MHeader(reader)
	if err != nil {
		decode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("cannot read decoded video header: %w%s", err, formatProcessStderr(<-decodeStderr))
	}
	if yinfo.chroma != "" && yinfo.chroma != "444" && !strings.HasPrefix(yinfo.chroma, "444") {
		decode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("unsupported decoded pixel type %q", yinfo.chroma)
	}
	outputHeight := yinfo.height * 3 / 4
	if yinfo.width <= 0 || outputHeight <= 0 {
		decode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("invalid decoded dimensions %dx%d", yinfo.width, yinfo.height)
	}

	encode := exec.Command(ffmpeg,
		"-y",
		"-v", "error",
		"-f", "rawvideo",
		"-pix_fmt", "bgra",
		"-s:v", fmt.Sprintf("%dx%d", yinfo.width, outputHeight),
		"-r", yinfo.frameRate,
		"-i", "-",
		"-an",
		"-c:v", "rawvideo",
		"-pix_fmt", "bgra",
		outputPath,
	)
	encodeIn, err := encode.StdinPipe()
	if err != nil {
		decode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("ffmpeg encode stdin: %w", err)
	}
	encodeErr, err := encode.StderrPipe()
	if err != nil {
		decode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("ffmpeg encode stderr: %w", err)
	}
	encodeStderr := capturePipe(encodeErr)
	if err := encode.Start(); err != nil {
		decode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("cannot start ffmpeg encoder: %w", err)
	}

	frameSize := yinfo.width * yinfo.height * 3
	yuv := make([]byte, frameSize)
	frameCount := 0
	for {
		ok, err := readY4MFrame(reader, yuv)
		if err != nil {
			encodeIn.Close()
			decode.Wait()
			encode.Wait()
			return 0, y4mInfo{}, err
		}
		if !ok {
			break
		}
		bgra, err := siglusAlphaYUV444ToBGRA(yuv, yinfo.width, yinfo.height)
		if err != nil {
			encodeIn.Close()
			decode.Wait()
			encode.Wait()
			return 0, y4mInfo{}, err
		}
		if _, err := encodeIn.Write(bgra); err != nil {
			encodeIn.Close()
			decode.Wait()
			encode.Wait()
			return 0, y4mInfo{}, fmt.Errorf("cannot feed ffmpeg encoder: %w%s", err, formatProcessStderr(<-encodeStderr))
		}
		frameCount++
	}
	if err := encodeIn.Close(); err != nil {
		decode.Wait()
		encode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("cannot close ffmpeg encoder input: %w", err)
	}
	if err := decode.Wait(); err != nil {
		encode.Wait()
		return 0, y4mInfo{}, fmt.Errorf("ffmpeg decoder failed: %w%s", err, formatProcessStderr(<-decodeStderr))
	}
	if err := encode.Wait(); err != nil {
		return 0, y4mInfo{}, fmt.Errorf("ffmpeg encoder failed: %w%s", err, formatProcessStderr(<-encodeStderr))
	}
	return frameCount, yinfo, nil
}

func resolveFFmpeg(path string) (string, error) {
	if strings.TrimSpace(path) != "" {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("ffmpeg not found at %s: %w", path, err)
		}
		return path, nil
	}
	if env := strings.TrimSpace(os.Getenv("SIGLUS_FFMPEG")); env != "" {
		if _, err := os.Stat(env); err == nil {
			return env, nil
		}
	}
	if found, err := exec.LookPath("ffmpeg"); err == nil {
		return found, nil
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, candidate := range []string{
			filepath.Join(dir, "ffmpeg.exe"),
			filepath.Join(dir, "tools", "ffmpeg.exe"),
		} {
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("ffmpeg not found; install ffmpeg, put ffmpeg.exe next to siglustest.exe, or set SIGLUS_FFMPEG")
}

func readY4MHeader(reader *bufio.Reader) (y4mInfo, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return y4mInfo{}, err
	}
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) == 0 || parts[0] != "YUV4MPEG2" {
		return y4mInfo{}, fmt.Errorf("invalid YUV4MPEG stream")
	}
	info := y4mInfo{frameRate: "30/1"}
	for _, part := range parts[1:] {
		if len(part) < 2 {
			continue
		}
		switch part[0] {
		case 'W':
			info.width, _ = strconv.Atoi(part[1:])
		case 'H':
			info.height, _ = strconv.Atoi(part[1:])
		case 'F':
			info.frameRate = strings.ReplaceAll(part[1:], ":", "/")
		case 'C':
			info.chroma = part[1:]
		}
	}
	if info.width <= 0 || info.height <= 0 {
		return y4mInfo{}, fmt.Errorf("missing YUV4MPEG dimensions")
	}
	return info, nil
}

func readY4MFrame(reader *bufio.Reader, dst []byte) (bool, error) {
	header, err := reader.ReadString('\n')
	if err == io.EOF {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cannot read frame header: %w", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(header), "FRAME") {
		return false, fmt.Errorf("invalid frame header %q", strings.TrimSpace(header))
	}
	if _, err := io.ReadFull(reader, dst); err != nil {
		return false, fmt.Errorf("cannot read decoded frame: %w", err)
	}
	return true, nil
}

func siglusAlphaYUV444ToBGRA(yuv []byte, width, encodedHeight int) ([]byte, error) {
	if width <= 0 || encodedHeight <= 0 {
		return nil, fmt.Errorf("invalid frame dimensions")
	}
	outputHeight := encodedHeight * 3 / 4
	if outputHeight <= 0 {
		return nil, fmt.Errorf("invalid alpha frame height")
	}
	required := width * encodedHeight * 3
	if len(yuv) < required {
		return nil, fmt.Errorf("decoded frame is too small")
	}
	out := make([]byte, width*outputHeight*4)
	third := (outputHeight + 2) / 3
	for y := 0; y < outputHeight; y++ {
		for x := 0; x < width; x++ {
			dst := (y*width + x) * 4
			out[dst+0] = yuv[width*(encodedHeight*0+y)+x]
			out[dst+1] = yuv[width*(encodedHeight*1+y)+x]
			out[dst+2] = yuv[width*(encodedHeight*2+y)+x]
			alphaRow := outputHeight + y
			if y >= third && y < third*2 {
				alphaRow = outputHeight*2 + y
			} else if y >= third*2 {
				alphaRow = outputHeight*3 + y
			}
			alpha := width*alphaRow + x
			if alpha < 0 || alpha >= len(yuv) {
				return nil, fmt.Errorf("alpha plane index out of bounds")
			}
			out[dst+3] = yuv[alpha]
		}
	}
	return out, nil
}

func replaceExtension(path, ext string) string {
	old := filepath.Ext(path)
	if old == "" {
		return path + ext
	}
	return strings.TrimSuffix(path, old) + ext
}

func formatProcessStderr(stderr string) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return ""
	}
	return ": " + stderr
}

func capturePipe(reader io.Reader) <-chan string {
	ch := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, reader)
		ch <- buf.String()
	}()
	return ch
}
