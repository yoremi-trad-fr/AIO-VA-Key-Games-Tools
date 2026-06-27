package siglus

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestCutOMVHeader(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "input.omv")
	outPath := filepath.Join(dir, "output.ogv")
	expected := []byte{'O', 'g', 'g', 'S', 0x01, 0x02, 0x03}
	input := append([]byte{0x10, 0x20, 0x30}, expected...)
	if err := os.WriteFile(inPath, input, 0644); err != nil {
		t.Fatal(err)
	}
	if err := CutOMVHeader(inPath, outPath); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, expected) {
		t.Fatalf("unexpected OGV payload: % X", got)
	}
}

func TestPackOMVBuildsHeaderAndIndexes(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "input.ogv")
	outPath := filepath.Join(dir, "output.omv")
	ogv := makeTestTheoraOGV()
	if err := os.WriteFile(inPath, ogv, 0644); err != nil {
		t.Fatal(err)
	}
	if err := PackOMV(inPath, outPath); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	oggOffset := bytes.Index(got, []byte("OggS"))
	const wantOggOffset = 168 + 4*28 + 2*32
	if oggOffset != wantOggOffset {
		t.Fatalf("unexpected Ogg offset: got %d, want %d", oggOffset, wantOggOffset)
	}
	if !bytes.Equal(got[oggOffset:], ogv) {
		t.Fatalf("embedded OGV changed")
	}

	readU32 := func(pos int) uint32 {
		t.Helper()
		return binary.LittleEndian.Uint32(got[pos : pos+4])
	}
	readI32 := func(pos int) int32 {
		t.Helper()
		return int32(readU32(pos))
	}

	if readU32(0) != 168 || readU32(4) != 0x101 {
		t.Fatalf("unexpected OMV header prefix")
	}
	if readU32(0x28) != 2 {
		t.Fatalf("unexpected OMV color type: %d", readU32(0x28))
	}
	if readU32(0x2c) != 320 || readU32(0x30) != 240 {
		t.Fatalf("unexpected dimensions: %dx%d", readU32(0x2c), readU32(0x30))
	}
	if readU32(0x3c) != 33333 {
		t.Fatalf("unexpected frame time: %d", readU32(0x3c))
	}
	if readU32(0x4c) != 4 || readU32(0x50) != 2 {
		t.Fatalf("unexpected index counts: pages=%d frames=%d", readU32(0x4c), readU32(0x50))
	}

	pageIndex := 168
	if readI32(pageIndex+20) != -1 || readI32(pageIndex+28+20) != -1 {
		t.Fatalf("first two pages should be marked as header pages")
	}
	page3 := pageIndex + 3*28
	if readU32(page3) != 3 || readI32(page3+20) != 2 || readI32(page3+24) != 0 {
		t.Fatalf("unexpected frame stats for page 3")
	}

	frameIndex := 168 + 4*28
	if readU32(frameIndex) != 0 || readU32(frameIndex+4) != 3 || readU32(frameIndex+12) != 1 {
		t.Fatalf("unexpected first frame index")
	}
	secondFrame := frameIndex + 32
	if readU32(secondFrame) != 1 || readU32(secondFrame+4) != 3 || readU32(secondFrame+8) != 1 {
		t.Fatalf("unexpected second frame position")
	}
	if readU32(secondFrame+16) != 0 || readU32(secondFrame+20) != 3 {
		t.Fatalf("unexpected keyframe backlink")
	}
	if readU32(secondFrame+24) != 34 || readU32(secondFrame+28) != 66 {
		t.Fatalf("unexpected second frame timing: %d-%d", readU32(secondFrame+24), readU32(secondFrame+28))
	}
}

func makeTestTheoraOGV() []byte {
	const serial = 0x12345678
	var out []byte
	out = append(out, makeOggPage(0, 0x02, serial, makeTheoraIdentificationPacket())...)
	out = append(out, makeOggPage(1, 0x00, serial, []byte{0x81, 't', 'h', 'e', 'o', 'r', 'a'})...)
	out = append(out, makeOggPage(2, 0x00, serial, []byte{0x82, 't', 'h', 'e', 'o', 'r', 'a'})...)
	out = append(out, makeOggPage(3, 0x00, serial, []byte{0x00, 0x01}, []byte{0x40, 0x02})...)
	return out
}

func makeOggPage(sequence uint32, headerType byte, serial uint32, packets ...[]byte) []byte {
	segmentCount := len(packets)
	bodyLen := 0
	for _, packet := range packets {
		bodyLen += len(packet)
	}
	page := make([]byte, 27+segmentCount, 27+segmentCount+bodyLen)
	copy(page[:4], []byte("OggS"))
	page[5] = headerType
	binary.LittleEndian.PutUint32(page[14:18], serial)
	binary.LittleEndian.PutUint32(page[18:22], sequence)
	page[26] = byte(segmentCount)
	for i, packet := range packets {
		page[27+i] = byte(len(packet))
	}
	for _, packet := range packets {
		page = append(page, packet...)
	}
	return page
}

func makeTheoraIdentificationPacket() []byte {
	packet := []byte{0x80, 't', 'h', 'e', 'o', 'r', 'a'}
	var bw testMSBBitWriter
	bw.write(3, 8)
	bw.write(2, 8)
	bw.write(1, 8)
	bw.write(20, 16)  // 320px encoded frame width
	bw.write(15, 16)  // 240px encoded frame height
	bw.write(320, 24) // picture width
	bw.write(240, 24) // picture height
	bw.write(0, 8)
	bw.write(0, 8)
	bw.write(30, 32) // fps numerator
	bw.write(1, 32)  // fps denominator
	bw.write(1, 24)
	bw.write(1, 24)
	bw.write(0, 8)
	bw.write(0, 24)
	bw.write(0, 6)
	bw.write(6, 5)
	bw.write(3, 2)
	bw.write(0, 3)
	return append(packet, bw.data...)
}

type testMSBBitWriter struct {
	data []byte
	bit  int
}

func (w *testMSBBitWriter) write(value uint32, bits int) {
	for i := bits - 1; i >= 0; i-- {
		if w.bit == 0 {
			w.data = append(w.data, 0)
		}
		if value&(1<<i) != 0 {
			w.data[len(w.data)-1] |= 1 << (7 - w.bit)
		}
		w.bit = (w.bit + 1) % 8
	}
}
