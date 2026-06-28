package siglus

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func TestRepackScriptTextReplacesSequentialLinesAndKeepsRest(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "test.ss")
	textPath := filepath.Join(dir, "lines.txt")
	outputPath := filepath.Join(dir, "patched.ss")

	orig := makeTestScript(t, []string{"old0", "old1", "old2"})
	if err := os.WriteFile(scriptPath, orig, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(textPath, makeUTF16LEText("new0\r\n新1"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RepackScriptText(scriptPath, textPath, outputPath); err != nil {
		t.Fatal(err)
	}

	gotLines, err := DumpSS(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"new0", "新1", "old2"}
	if len(gotLines) != len(want) {
		t.Fatalf("line count = %d, want %d", len(gotLines), len(want))
	}
	for i, line := range gotLines {
		if line.Text != want[i] {
			t.Fatalf("line %d = %q, want %q", i, line.Text, want[i])
		}
	}

	out, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	newTableOffset := binary.LittleEndian.Uint32(out[4+2*8:])
	if int(newTableOffset) != len(orig) {
		t.Fatalf("new string table offset = %d, want %d", newTableOffset, len(orig))
	}
}

func TestRepackScriptTextRejectsTooManyLines(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "test.ss")
	textPath := filepath.Join(dir, "lines.txt")
	if err := os.WriteFile(scriptPath, makeTestScript(t, []string{"old0"}), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(textPath, makeUTF16LEText("a\r\nb"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := RepackScriptText(scriptPath, textPath, ""); err == nil {
		t.Fatalf("expected too many lines error")
	}
}

func makeTestScript(t *testing.T, strings []string) []byte {
	t.Helper()
	idxOffset := ssHeaderSize
	tblOffset := idxOffset + len(strings)*8
	buf := make([]byte, tblOffset)
	binary.LittleEndian.PutUint32(buf[0:], uint32(ssHeaderSize))
	binary.LittleEndian.PutUint32(buf[4+1*8:], uint32(idxOffset))
	binary.LittleEndian.PutUint32(buf[4+1*8+4:], uint32(len(strings)))
	binary.LittleEndian.PutUint32(buf[4+2*8:], uint32(tblOffset))

	currentChars := 0
	var table []byte
	for i, text := range strings {
		encoded := encryptString(text, i)
		entryOff := idxOffset + i*8
		binary.LittleEndian.PutUint32(buf[entryOff:], uint32(currentChars))
		binary.LittleEndian.PutUint32(buf[entryOff+4:], uint32(len(encoded)))
		for _, v := range encoded {
			var b [2]byte
			binary.LittleEndian.PutUint16(b[:], v)
			table = append(table, b[:]...)
		}
		currentChars += len(encoded)
	}
	binary.LittleEndian.PutUint32(buf[4+2*8+4:], uint32(currentChars))
	return append(buf, table...)
}

func makeUTF16LEText(text string) []byte {
	out := []byte{0xff, 0xfe}
	for _, v := range utf16.Encode([]rune(text)) {
		var b [2]byte
		binary.LittleEndian.PutUint16(b[:], v)
		out = append(out, b[:]...)
	}
	return out
}
