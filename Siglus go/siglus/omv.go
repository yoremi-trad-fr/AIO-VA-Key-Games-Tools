package siglus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// CutOMVHeader writes the embedded Ogg stream from a Siglus OMV file.
func CutOMVHeader(inputPath, outputPath string) error {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read OMV: %w", err)
	}
	offset := bytes.Index(buf, []byte("OggS"))
	if offset < 0 {
		return fmt.Errorf("OggS marker not found")
	}
	if err := os.WriteFile(outputPath, buf[offset:], 0644); err != nil {
		return fmt.Errorf("cannot write OGV: %w", err)
	}
	return nil
}

// PackOMV wraps an Ogg/Theora stream in the Siglus OMV 1.1 container.
func PackOMV(inputPath, outputPath string) error {
	ogv, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read OGV: %w", err)
	}
	packed, err := buildOMV(ogv)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, packed, 0644); err != nil {
		return fmt.Errorf("cannot write OMV: %w", err)
	}
	return nil
}

type OMVBuildOptions struct {
	EncryptType uint32
	Width       uint32
	Height      uint32
}

type omvOggPage struct {
	offset     uint32
	length     uint32
	headerType byte
	serial     uint32
	sequence   uint32
	segments   []byte
	body       []byte
}

type omvPageIndex struct {
	idx        uint32
	wtf        uint32
	length     uint32
	offset     uint32
	idxStart   int32
	numFrames  int32
	frameStart int32
}

type omvFrameIndex struct {
	frame           uint32
	pageIdx         uint32
	seqCount        uint32
	wtf             uint32
	keyframe        uint32
	keyframePageIdx uint32
	start           uint32
	end             uint32
}

type theoraInfo struct {
	frameWidth     uint32
	frameHeight    uint32
	fpsNumerator   uint32
	fpsDenominator uint32
	pixelFormat    uint32
}

func buildOMV(ogv []byte) ([]byte, error) {
	return buildOMVWithOptions(ogv, OMVBuildOptions{})
}

func buildOMVWithOptions(ogv []byte, opts OMVBuildOptions) ([]byte, error) {
	pages, err := parseOggPages(ogv)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("OGV contains no Ogg pages")
	}

	serial, info, err := findTheoraStream(pages)
	if err != nil {
		return nil, err
	}
	pageIdx, frameIdx, err := indexTheoraFrames(pages, serial, info)
	if err != nil {
		return nil, err
	}

	if opts.EncryptType == 0 {
		opts.EncryptType = 2
	}
	if opts.Width == 0 {
		opts.Width = info.frameWidth
	}
	if opts.Height == 0 {
		opts.Height = info.frameHeight
	}

	headerLen := 168
	out := bytes.NewBuffer(make([]byte, 0, headerLen+len(pageIdx)*28+len(frameIdx)*32+len(ogv)))
	writeOMVHeader(out, info, serial, uint32(len(pageIdx)), uint32(len(frameIdx)), opts)
	for _, p := range pageIdx {
		writeU32(out, p.idx)
		writeU32(out, p.wtf)
		writeU32(out, p.length)
		writeU32(out, p.offset)
		writeI32(out, p.idxStart)
		writeI32(out, p.numFrames)
		writeI32(out, p.frameStart)
	}
	for _, f := range frameIdx {
		writeU32(out, f.frame)
		writeU32(out, f.pageIdx)
		writeU32(out, f.seqCount)
		writeU32(out, f.wtf)
		writeU32(out, f.keyframe)
		writeU32(out, f.keyframePageIdx)
		writeU32(out, f.start)
		writeU32(out, f.end)
	}
	out.Write(ogv)
	return out.Bytes(), nil
}

func parseOggPages(data []byte) ([]omvOggPage, error) {
	var pages []omvOggPage
	for pos := 0; pos < len(data); {
		if pos+27 > len(data) {
			return nil, fmt.Errorf("truncated Ogg page at 0x%x", pos)
		}
		if !bytes.Equal(data[pos:pos+4], []byte("OggS")) {
			return nil, fmt.Errorf("OggS marker not found at 0x%x", pos)
		}
		if data[pos+4] != 0 {
			return nil, fmt.Errorf("unsupported Ogg stream structure version %d", data[pos+4])
		}
		segmentCount := int(data[pos+26])
		headerLen := 27 + segmentCount
		if pos+headerLen > len(data) {
			return nil, fmt.Errorf("truncated Ogg segment table at 0x%x", pos)
		}
		segments := data[pos+27 : pos+headerLen]
		bodyLen := 0
		for _, seg := range segments {
			bodyLen += int(seg)
		}
		pageLen := headerLen + bodyLen
		if pos+pageLen > len(data) {
			return nil, fmt.Errorf("truncated Ogg page body at 0x%x", pos)
		}
		pages = append(pages, omvOggPage{
			offset:     uint32(pos),
			length:     uint32(pageLen),
			headerType: data[pos+5],
			serial:     binary.LittleEndian.Uint32(data[pos+14 : pos+18]),
			sequence:   binary.LittleEndian.Uint32(data[pos+18 : pos+22]),
			segments:   segments,
			body:       data[pos+headerLen : pos+pageLen],
		})
		pos += pageLen
	}
	return pages, nil
}

func findTheoraStream(pages []omvOggPage) (uint32, theoraInfo, error) {
	pending := make(map[uint32][]byte)
	for _, page := range pages {
		bodyPos := 0
		packet := pending[page.serial]
		if page.headerType&0x01 == 0 {
			packet = packet[:0]
		}
		for _, seg := range page.segments {
			next := bodyPos + int(seg)
			packet = append(packet, page.body[bodyPos:next]...)
			bodyPos = next
			if seg < 255 {
				if isTheoraIdentification(packet) {
					info, err := parseTheoraIdentification(packet)
					if err != nil {
						return 0, theoraInfo{}, err
					}
					return page.serial, info, nil
				}
				packet = packet[:0]
			}
		}
		if len(packet) > 0 {
			pending[page.serial] = packet
		} else {
			delete(pending, page.serial)
		}
	}
	return 0, theoraInfo{}, fmt.Errorf("Theora stream not found")
}

func indexTheoraFrames(pages []omvOggPage, serial uint32, info theoraInfo) ([]omvPageIndex, []omvFrameIndex, error) {
	pageIdx := make([]omvPageIndex, len(pages))
	for i, page := range pages {
		if page.sequence != uint32(i) {
			return nil, nil, fmt.Errorf("unexpected Ogg page sequence: got %d, expected %d", page.sequence, i)
		}
		pageIdx[i] = omvPageIndex{
			idx:        uint32(i),
			wtf:        256,
			length:     page.length,
			offset:     page.offset,
			idxStart:   0,
			numFrames:  0,
			frameStart: -1,
		}
	}

	var frames []omvFrameIndex
	var packet []byte
	var keyframe uint32
	var keyframePage uint32
	frameTimeMS := float32(info.fpsDenominator) / float32(info.fpsNumerator) * 1000.0
	var prevEnd uint32

	for pageNumber, page := range pages {
		if page.serial != serial {
			continue
		}
		bodyPos := 0
		if page.headerType&0x01 == 0 {
			packet = packet[:0]
		}
		for _, seg := range page.segments {
			next := bodyPos + int(seg)
			packet = append(packet, page.body[bodyPos:next]...)
			bodyPos = next
			if seg == 255 {
				continue
			}

			if isTheoraDataPacket(packet) {
				frame := uint32(len(frames))
				if isTheoraKeyframe(packet) {
					keyframe = frame
					keyframePage = uint32(pageNumber)
				}

				pageEntry := &pageIdx[pageNumber]
				if pageEntry.frameStart < 0 {
					pageEntry.frameStart = int32(frame)
				}
				seqCount := uint32(pageEntry.numFrames)
				pageEntry.numFrames++

				start := uint32(math.Round(float64(float32(frame) * frameTimeMS)))
				end := uint32(math.Round(float64(float32(start) + frameTimeMS)))
				if frame > 0 && start == prevEnd {
					start++
				}

				flags := uint32(0)
				if frame == keyframe {
					flags = 1
				}
				frames = append(frames, omvFrameIndex{
					frame:           frame,
					pageIdx:         uint32(pageNumber),
					seqCount:        seqCount,
					wtf:             flags,
					keyframe:        keyframe,
					keyframePageIdx: keyframePage,
					start:           start,
					end:             end,
				})
				prevEnd = end
			}
			packet = packet[:0]
		}
	}

	if len(frames) == 0 {
		return nil, nil, fmt.Errorf("no Theora video frames found")
	}
	if len(pageIdx) > 0 {
		pageIdx[0].numFrames = -1
	}
	if len(pageIdx) > 1 {
		pageIdx[1].numFrames = -1
	}
	return pageIdx, frames, nil
}

func writeOMVHeader(out *bytes.Buffer, info theoraInfo, serial uint32, pageCount, frameCount uint32, opts OMVBuildOptions) {
	frameTimeUS := uint32(float32(info.fpsDenominator) / float32(info.fpsNumerator) * 1000000.0)

	writeU32(out, 168)
	writeU32(out, 0x101)
	for i := 0; i < 8; i++ {
		writeU32(out, 0)
	}
	writeU32(out, opts.EncryptType)
	writeU32(out, opts.Width)
	writeU32(out, opts.Height)
	writeU32(out, 0)
	writeU32(out, 0)
	writeU32(out, frameTimeUS)
	writeU32(out, serial)
	writeU32(out, 0)
	writeU32(out, 1)
	writeU32(out, pageCount)
	writeU32(out, frameCount)
	for i := 0; i < 7; i++ {
		writeU32(out, 0)
	}
	writeU32(out, 0)
	writeU32(out, 0)
	writeI32(out, -1)
	writeI32(out, -1)
	writeU32(out, 0)
	writeU32(out, 0)
	writeU32(out, 0)
	for i := 0; i < 7; i++ {
		writeU32(out, 0)
	}
}

func isTheoraIdentification(packet []byte) bool {
	return len(packet) >= 7 && packet[0] == 0x80 && string(packet[1:7]) == "theora"
}

func isTheoraDataPacket(packet []byte) bool {
	return len(packet) == 0 || packet[0]&0x80 == 0
}

func isTheoraKeyframe(packet []byte) bool {
	return len(packet) > 0 && packet[0]&0x80 == 0 && packet[0]&0x40 == 0
}

func parseTheoraIdentification(packet []byte) (theoraInfo, error) {
	if !isTheoraIdentification(packet) {
		return theoraInfo{}, fmt.Errorf("not a Theora identification packet")
	}
	br := newMSBBitReader(packet[7:])
	br.read(8) // version major
	br.read(8) // version minor
	br.read(8) // version subminor
	frameWidthMB := br.read(16)
	frameHeightMB := br.read(16)
	br.read(24) // picture width
	br.read(24) // picture height
	br.read(8)  // picture x
	br.read(8)  // picture y
	fpsNumerator := br.read(32)
	fpsDenominator := br.read(32)
	br.read(24) // aspect numerator
	br.read(24) // aspect denominator
	br.read(8)  // colorspace
	br.read(24) // target bitrate
	br.read(6)  // quality
	br.read(5)  // keyframe granule shift
	pixelFormat := br.read(2)
	br.read(3) // reserved
	if br.err != nil {
		return theoraInfo{}, br.err
	}
	if frameWidthMB == 0 || frameHeightMB == 0 || fpsNumerator == 0 || fpsDenominator == 0 {
		return theoraInfo{}, fmt.Errorf("invalid Theora identification header")
	}
	if pixelFormat == 1 || pixelFormat >= 4 {
		return theoraInfo{}, fmt.Errorf("unsupported Theora pixel format %d", pixelFormat)
	}
	return theoraInfo{
		frameWidth:     frameWidthMB << 4,
		frameHeight:    frameHeightMB << 4,
		fpsNumerator:   fpsNumerator,
		fpsDenominator: fpsDenominator,
		pixelFormat:    pixelFormat,
	}, nil
}

type msbBitReader struct {
	data []byte
	bit  int
	err  error
}

func newMSBBitReader(data []byte) *msbBitReader {
	return &msbBitReader{data: data}
}

func (r *msbBitReader) read(bits int) uint32 {
	if r.err != nil {
		return 0
	}
	if bits < 0 || bits > 32 || r.bit+bits > len(r.data)*8 {
		r.err = fmt.Errorf("truncated Theora identification header")
		return 0
	}
	var v uint32
	for i := 0; i < bits; i++ {
		byteIndex := (r.bit + i) / 8
		bitIndex := 7 - ((r.bit + i) % 8)
		v = (v << 1) | uint32((r.data[byteIndex]>>bitIndex)&1)
	}
	r.bit += bits
	return v
}

func writeU32(out *bytes.Buffer, v uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	out.Write(buf[:])
}

func writeI32(out *bytes.Buffer, v int32) {
	writeU32(out, uint32(v))
}
