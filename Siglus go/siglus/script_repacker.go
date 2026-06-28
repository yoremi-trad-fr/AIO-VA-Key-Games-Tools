package siglus

import (
	"encoding/binary"
	"fmt"
	"os"
	"unicode/utf16"
)

func DefaultScriptRepackOutput(scriptPath string) string {
	return scriptPath + ".out"
}

func RepackScriptText(scriptPath, textPath, outputPath string) error {
	if outputPath == "" {
		outputPath = DefaultScriptRepackOutput(scriptPath)
	}
	lines, err := readUTF16TextLines(textPath)
	if err != nil {
		return err
	}
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("cannot read script: %w", err)
	}
	if len(script) < ssHeaderSize {
		return fmt.Errorf("file too small to be a valid script")
	}

	hdr := readSSHeader(script)
	idxOffset := int(hdr.StrIndex.Offset)
	idxCount := int(hdr.StrIndex.Size)
	tblOffset := int(hdr.StrTable.Offset)
	if idxOffset < 0 || tblOffset < 0 || idxOffset+idxCount*8 > len(script) {
		return fmt.Errorf("string index out of bounds")
	}
	if len(lines) > idxCount {
		return fmt.Errorf("index error: text has %d lines but script has %d strings", len(lines), idxCount)
	}

	newTable := make([]byte, 0)
	for i := 0; i < idxCount; i++ {
		text := ""
		if i < len(lines) {
			text = lines[i]
		} else {
			entryOff := idxOffset + i*8
			charOffset := int(binary.LittleEndian.Uint32(script[entryOff:]))
			charSize := int(binary.LittleEndian.Uint32(script[entryOff+4:]))
			text, err = readDecryptedSSString(script, tblOffset, charOffset, charSize, i)
			if err != nil {
				return err
			}
		}

		encoded := encryptString(text, i)
		entryOff := idxOffset + i*8
		binary.LittleEndian.PutUint32(script[entryOff:], uint32(len(newTable)/2))
		binary.LittleEndian.PutUint32(script[entryOff+4:], uint32(len(encoded)))
		for _, v := range encoded {
			var b [2]byte
			binary.LittleEndian.PutUint16(b[:], v)
			newTable = append(newTable, b[:]...)
		}
	}

	out := make([]byte, 0, len(script)+len(newTable))
	out = append(out, script...)
	binary.LittleEndian.PutUint32(out[4+2*8:], uint32(len(script)))
	out = append(out, newTable...)
	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		return fmt.Errorf("cannot write repacked script: %w", err)
	}
	return nil
}

func readDecryptedSSString(buf []byte, tblOffset, charOffset, charSize, index int) (string, error) {
	if charSize == 0 {
		return "", nil
	}
	byteOffset := tblOffset + charOffset*2
	byteSize := charSize * 2
	if byteOffset < 0 || byteOffset+byteSize > len(buf) {
		return "", fmt.Errorf("string %d out of bounds", index)
	}
	u16 := make([]uint16, charSize)
	key := uint16(index * 0x7087)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(buf[byteOffset+i*2:]) ^ key
	}
	return string(utf16.Decode(u16)), nil
}

func readUTF16TextLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read text file: %w", err)
	}
	if len(data) < 2 {
		return nil, fmt.Errorf("invalid text file size")
	}

	bigEndian := data[0] == 0xfe && data[1] == 0xff
	body := data[2:]
	if len(body)%2 != 0 {
		return nil, fmt.Errorf("UTF-16 text has odd byte length")
	}
	u16 := make([]uint16, len(body)/2)
	for i := range u16 {
		if bigEndian {
			u16[i] = binary.BigEndian.Uint16(body[i*2:])
		} else {
			u16[i] = binary.LittleEndian.Uint16(body[i*2:])
		}
	}
	runes := []rune(string(utf16.Decode(u16)))
	var lines []string
	start := 0
	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '\r':
			lines = append(lines, string(runes[start:i]))
			if i+1 < len(runes) && runes[i+1] == '\n' {
				i++
			}
			start = i + 1
		case '\n':
			lines = append(lines, string(runes[start:i]))
			start = i + 1
		}
	}
	if start < len(runes) {
		lines = append(lines, string(runes[start:]))
	}
	return lines, nil
}
