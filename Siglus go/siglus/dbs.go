package siglus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"

	textencoding "golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	dbsTypeString = 0x53
	dbsTypeInt    = 0x56
	dbsPadSize    = 64
)

// DBSPackOptions controls DBS rebuilding.
// ANSIEncoding defaults to GBK to match Siglus Tools 0.61 dbsEncrypt.py.
type DBSPackOptions struct {
	CompressionLevel int
	FakeCompression  bool
	ANSIEncoding     string
}

type dbsHeader struct {
	fileSize            uint32
	headerData          []byte
	lineCount           uint32
	dataCount           uint32
	lineIndexOffset     uint32
	dataFormatOffset    uint32
	lineDataIndexOffset uint32
	textOffset          uint32
}

type dbsColumn struct {
	index uint32
	typ   uint32
}

type dbsValue struct {
	text string
	num  uint32
}

type dbsTable struct {
	header    dbsHeader
	isUnicode bool
	lineIndex []int32
	columns   []dbsColumn
	rows      [][]dbsValue
}

// UnpackDBS decrypts a .dbs file and writes both the raw .dbs.out and UTF-16 text dump.
func UnpackDBS(inputPath, rawOutputPath, txtOutputPath string, dumpAll bool) error {
	return UnpackDBSWithEncoding(inputPath, rawOutputPath, txtOutputPath, dumpAll, "shift-jis")
}

// UnpackDBSWithEncoding is UnpackDBS with an explicit ANSI decoder for non-Unicode DBS files.
func UnpackDBSWithEncoding(inputPath, rawOutputPath, txtOutputPath string, dumpAll bool, ansiEncoding string) error {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read DBS: %w", err)
	}
	if len(buf) < 12 {
		return fmt.Errorf("file too small to be a DBS")
	}

	isUnicode := binary.LittleEndian.Uint32(buf[0:4]) != 0
	data := append([]byte(nil), buf[4:]...)
	xorDBS(data)

	decompressed := Decompress(data)
	if decompressed == nil {
		return fmt.Errorf("DBS decompression failed")
	}
	xorDBSLayer(decompressed)

	if rawOutputPath != "" {
		if err := os.WriteFile(rawOutputPath, decompressed, 0644); err != nil {
			return fmt.Errorf("cannot write DBS raw output: %w", err)
		}
	}

	table, err := parseDBSRaw(decompressed, isUnicode, ansiEncoding)
	if err != nil {
		return err
	}
	txt := formatDBSText(table, dumpAll)
	if err := os.WriteFile(txtOutputPath, encodeUTF16LEWithBOM(txt), 0644); err != nil {
		return fmt.Errorf("cannot write DBS text output: %w", err)
	}
	return nil
}

// PackDBS rebuilds a .dbs from a raw .dbs.out and a UTF-16 text dump.
func PackDBS(rawOutPath, txtPath, outputPath string, compressionLevel int, fakeCompression bool) error {
	return PackDBSWithOptions(rawOutPath, txtPath, outputPath, DBSPackOptions{
		CompressionLevel: compressionLevel,
		FakeCompression:  fakeCompression,
		ANSIEncoding:     "gbk",
	})
}

// PackDBSWithOptions rebuilds a .dbs with explicit compression and ANSI encoding settings.
func PackDBSWithOptions(rawOutPath, txtPath, outputPath string, opts DBSPackOptions) error {
	raw, err := os.ReadFile(rawOutPath)
	if err != nil {
		return fmt.Errorf("cannot read DBS raw output: %w", err)
	}
	txtBytes, err := os.ReadFile(txtPath)
	if err != nil {
		return fmt.Errorf("cannot read DBS text: %w", err)
	}
	txt, err := decodeUTF16Text(txtBytes)
	if err != nil {
		return fmt.Errorf("cannot decode DBS text: %w", err)
	}
	isUnicode, err := dbsTextMode(txt)
	if err != nil {
		return err
	}

	table, err := parseDBSRaw(raw, isUnicode, "shift-jis")
	if err != nil {
		return err
	}
	applyDBSText(table, txt)

	if opts.ANSIEncoding == "" {
		opts.ANSIEncoding = "gbk"
	}
	rebuiltRaw, err := buildDBSRaw(table, opts.ANSIEncoding)
	if err != nil {
		return err
	}
	xorDBSLayer(rebuiltRaw)

	var packed []byte
	if opts.FakeCompression {
		packed = FakeCompress(rebuiltRaw)
	} else {
		packed = CompressLevel(rebuiltRaw, opts.CompressionLevel)
	}
	xorDBS(packed)

	out := make([]byte, 4, len(packed)+4)
	if isUnicode {
		binary.LittleEndian.PutUint32(out[0:4], 1)
	}
	out = append(out, packed...)
	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		return fmt.Errorf("cannot write DBS: %w", err)
	}
	return nil
}

func parseDBSRaw(raw []byte, isUnicode bool, ansiEncoding string) (*dbsTable, error) {
	header, err := readDBSHeader(raw)
	if err != nil {
		return nil, err
	}
	if header.fileSize == 0 || int(header.fileSize) > len(raw) {
		return nil, fmt.Errorf("DBS raw file size is invalid")
	}
	lineCount, dataCount := int(header.lineCount), int(header.dataCount)
	if lineCount < 0 || dataCount < 0 || lineCount > 1_000_000 || dataCount > 10_000 {
		return nil, fmt.Errorf("DBS table dimensions are invalid")
	}

	lineIndexStart, lineIndexEnd, err := checkedRange(raw, header.lineIndexOffset, uint32(lineCount*4), "line index")
	if err != nil {
		return nil, err
	}
	lineIndex := make([]int32, lineCount)
	for i := range lineIndex {
		lineIndex[i] = int32(binary.LittleEndian.Uint32(raw[lineIndexStart+i*4:]))
	}
	_ = lineIndexEnd

	formatStart, _, err := checkedRange(raw, header.dataFormatOffset, uint32(dataCount*8), "data format")
	if err != nil {
		return nil, err
	}
	columns := make([]dbsColumn, dataCount)
	for i := range columns {
		off := formatStart + i*8
		columns[i] = dbsColumn{
			index: binary.LittleEndian.Uint32(raw[off:]),
			typ:   binary.LittleEndian.Uint32(raw[off+4:]),
		}
	}

	cellCount := lineCount * dataCount
	dataStart, _, err := checkedRange(raw, header.lineDataIndexOffset, uint32(cellCount*4), "line data")
	if err != nil {
		return nil, err
	}
	rows := make([][]dbsValue, lineCount)
	for row := range rows {
		rows[row] = make([]dbsValue, dataCount)
		for col := range columns {
			valueOff := dataStart + (row*dataCount+col)*4
			value := binary.LittleEndian.Uint32(raw[valueOff:])
			if columns[col].typ == dbsTypeString {
				text, err := readDBSString(raw, header.textOffset+value, isUnicode, ansiEncoding)
				if err != nil {
					return nil, fmt.Errorf("line %d data %d: %w", row, col, err)
				}
				rows[row][col].text = text
			} else {
				rows[row][col].num = value
			}
		}
	}

	return &dbsTable{
		header:    header,
		isUnicode: isUnicode,
		lineIndex: lineIndex,
		columns:   columns,
		rows:      rows,
	}, nil
}

func readDBSHeader(raw []byte) (dbsHeader, error) {
	if len(raw) < 28 {
		return dbsHeader{}, fmt.Errorf("DBS raw data is too small")
	}
	h := dbsHeader{
		fileSize:   binary.LittleEndian.Uint32(raw[0:4]),
		headerData: append([]byte(nil), raw[4:28]...),
	}
	h.lineCount = binary.LittleEndian.Uint32(h.headerData[0:4])
	h.dataCount = binary.LittleEndian.Uint32(h.headerData[4:8])
	h.lineIndexOffset = binary.LittleEndian.Uint32(h.headerData[8:12])
	h.dataFormatOffset = binary.LittleEndian.Uint32(h.headerData[12:16])
	h.lineDataIndexOffset = binary.LittleEndian.Uint32(h.headerData[16:20])
	h.textOffset = binary.LittleEndian.Uint32(h.headerData[20:24])
	return h, nil
}

func checkedRange(buf []byte, offset, size uint32, label string) (int, int, error) {
	start := int(offset)
	length := int(size)
	if uint32(start) != offset || uint32(length) != size || start < 0 || length < 0 {
		return 0, 0, fmt.Errorf("%s offset is invalid", label)
	}
	end := start + length
	if end < start || end > len(buf) {
		return 0, 0, fmt.Errorf("%s is out of bounds", label)
	}
	return start, end, nil
}

func readDBSString(raw []byte, offset uint32, isUnicode bool, ansiEncoding string) (string, error) {
	start := int(offset)
	if uint32(start) != offset || start < 0 || start > len(raw) {
		return "", fmt.Errorf("string offset is out of bounds")
	}
	if isUnicode {
		end := start
		for end+1 < len(raw) {
			if binary.LittleEndian.Uint16(raw[end:]) == 0 {
				break
			}
			end += 2
		}
		if end+1 >= len(raw) {
			return "", fmt.Errorf("unterminated UTF-16 string")
		}
		return decodeUTF16LE(raw[start:end]), nil
	}

	end := start
	for end < len(raw) && raw[end] != 0 {
		end++
	}
	if end >= len(raw) {
		return "", fmt.Errorf("unterminated ANSI string")
	}
	decoder, err := ansiDecoder(ansiEncoding)
	if err != nil {
		return "", err
	}
	decoded, err := decoder.Bytes(raw[start:end])
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func formatDBSText(table *dbsTable, dumpAll bool) string {
	var sb strings.Builder
	if table.isUnicode {
		sb.WriteString("Unicode\n")
	} else {
		sb.WriteString("ASCII\n")
	}
	for row := range table.rows {
		fmt.Fprintf(&sb, "[%04d]\n", table.lineIndex[row])
		for col, column := range table.columns {
			value := table.rows[row][col]
			switch column.typ {
			case dbsTypeString:
				if value.text != "" || dumpAll {
					fmt.Fprintf(&sb, "○%02d○%s\n●%02d●%s\n\n", column.index, value.text, column.index, value.text)
				}
			case dbsTypeInt:
				if dumpAll {
					fmt.Fprintf(&sb, "{%02d}%d\n<%02d>%d\n\n", column.index, value.num, column.index, value.num)
				}
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func dbsTextMode(txt string) (bool, error) {
	first, _, _ := strings.Cut(normalizeNewlines(txt), "\n")
	switch strings.TrimSpace(first) {
	case "Unicode":
		return true, nil
	case "ASCII":
		return false, nil
	default:
		return false, fmt.Errorf("DBS text must start with Unicode or ASCII")
	}
}

func applyDBSText(table *dbsTable, txt string) {
	lines := strings.Split(normalizeNewlines(txt), "\n")
	if len(lines) == 0 {
		return
	}

	lineMap := make(map[int32]int, len(table.lineIndex))
	for i, index := range table.lineIndex {
		lineMap[index] = i
	}
	columnMap := make(map[uint32]int, len(table.columns))
	for i, column := range table.columns {
		columnMap[column.index] = i
	}

	currentRow := -1
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			indexText, ok := betweenMarkers(line, "[", "]")
			if !ok {
				continue
			}
			index, err := strconv.ParseInt(indexText, 10, 32)
			if err != nil {
				currentRow = -1
				continue
			}
			if row, ok := lineMap[int32(index)]; ok {
				currentRow = row
			} else {
				currentRow = -1
			}
			continue
		}
		if currentRow < 0 {
			continue
		}
		if strings.HasPrefix(line, "●") {
			indexText, value, ok := markerValue(line, "●")
			if !ok {
				continue
			}
			index, err := strconv.ParseUint(indexText, 10, 32)
			if err != nil {
				continue
			}
			if col, ok := columnMap[uint32(index)]; ok && table.columns[col].typ == dbsTypeString {
				table.rows[currentRow][col].text = value
			}
			continue
		}
		if strings.HasPrefix(line, "<") {
			indexText, ok := betweenMarkers(line, "<", ">")
			if !ok {
				continue
			}
			index, err := strconv.ParseUint(indexText, 10, 32)
			if err != nil {
				continue
			}
			col, ok := columnMap[uint32(index)]
			if !ok || table.columns[col].typ != dbsTypeInt {
				continue
			}
			valueText := line[strings.Index(line, ">")+1:]
			if value, err := parseUint32Text(valueText); err == nil {
				table.rows[currentRow][col].num = value
			}
		}
	}
}

func betweenMarkers(line, open, close string) (string, bool) {
	if !strings.HasPrefix(line, open) {
		return "", false
	}
	rest := strings.TrimPrefix(line, open)
	pos := strings.Index(rest, close)
	if pos < 0 {
		return "", false
	}
	return rest[:pos], true
}

func markerValue(line, marker string) (string, string, bool) {
	if !strings.HasPrefix(line, marker) {
		return "", "", false
	}
	rest := strings.TrimPrefix(line, marker)
	pos := strings.Index(rest, marker)
	if pos < 0 {
		return "", "", false
	}
	return rest[:pos], rest[pos+len(marker):], true
}

func parseUint32Text(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "-") {
		value, err := strconv.ParseInt(s, 10, 32)
		return uint32(value), err
	}
	value, err := strconv.ParseUint(s, 0, 32)
	return uint32(value), err
}

func buildDBSRaw(table *dbsTable, ansiEncoding string) ([]byte, error) {
	var out bytes.Buffer
	out.Write(make([]byte, 4))
	out.Write(table.header.headerData)

	for _, index := range table.lineIndex {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], uint32(index))
		out.Write(b[:])
	}
	for _, column := range table.columns {
		var b [8]byte
		binary.LittleEndian.PutUint32(b[0:4], column.index)
		binary.LittleEndian.PutUint32(b[4:8], column.typ)
		out.Write(b[:])
	}
	if out.Len() != int(table.header.lineDataIndexOffset) {
		return nil, fmt.Errorf("unsupported DBS layout: line data offset mismatch")
	}

	var textData []byte
	textOffset := uint32(0)
	for row := range table.rows {
		for col, column := range table.columns {
			var b [4]byte
			value := table.rows[row][col]
			if column.typ == dbsTypeString {
				encoded, err := encodeDBSString(value.text, table.isUnicode, ansiEncoding)
				if err != nil {
					return nil, err
				}
				binary.LittleEndian.PutUint32(b[:], textOffset)
				textData = append(textData, encoded...)
				textOffset += uint32(len(encoded))
			} else {
				binary.LittleEndian.PutUint32(b[:], value.num)
			}
			out.Write(b[:])
		}
	}
	if out.Len() != int(table.header.textOffset) {
		return nil, fmt.Errorf("unsupported DBS layout: text offset mismatch")
	}
	out.Write(textData)

	raw := out.Bytes()
	dbsSize := len(raw)
	binary.LittleEndian.PutUint32(raw[0:4], uint32(dbsSize))
	padLen := dbsPadSize - dbsSize%dbsPadSize
	raw = append(raw, make([]byte, padLen)...)
	return raw, nil
}

func encodeDBSString(s string, isUnicode bool, ansiEncoding string) ([]byte, error) {
	if isUnicode {
		out := stringToUTF16LE(s)
		out = append(out, 0, 0)
		return out, nil
	}
	out, err := encodeANSIReplacing(s, ansiEncoding)
	if err != nil {
		return nil, err
	}
	out = append(out, 0)
	return out, nil
}

func encodeANSIReplacing(s, name string) ([]byte, error) {
	var out []byte
	for _, r := range s {
		encoder, err := ansiEncoder(name)
		if err != nil {
			return nil, err
		}
		encoded, err := encoder.Bytes([]byte(string(r)))
		if err != nil {
			replacement := "·"
			if normalizedEncoding(name) == "shift-jis" {
				replacement = "?"
			}
			encoder, _ = ansiEncoder(name)
			encoded, err = encoder.Bytes([]byte(replacement))
			if err != nil {
				return nil, err
			}
		}
		out = append(out, encoded...)
	}
	return out, nil
}

func ansiDecoder(name string) (*textencoding.Decoder, error) {
	switch normalizedEncoding(name) {
	case "shift-jis":
		return japanese.ShiftJIS.NewDecoder(), nil
	case "gbk":
		return simplifiedchinese.GBK.NewDecoder(), nil
	default:
		return nil, fmt.Errorf("unsupported ANSI encoding: %s", name)
	}
}

func ansiEncoder(name string) (*textencoding.Encoder, error) {
	switch normalizedEncoding(name) {
	case "shift-jis":
		return japanese.ShiftJIS.NewEncoder(), nil
	case "gbk":
		return simplifiedchinese.GBK.NewEncoder(), nil
	default:
		return nil, fmt.Errorf("unsupported ANSI encoding: %s", name)
	}
}

func normalizedEncoding(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "_", "-")
	switch name {
	case "", "ansi", "gbk", "gb2312", "cp936":
		return "gbk"
	case "shift-jis", "shiftjis", "sjis", "cp932":
		return "shift-jis"
	default:
		return name
	}
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func decodeUTF16Text(data []byte) (string, error) {
	if len(data) < 2 {
		return "", fmt.Errorf("empty UTF-16 text")
	}
	bigEndian := false
	switch {
	case data[0] == 0xFF && data[1] == 0xFE:
		data = data[2:]
	case data[0] == 0xFE && data[1] == 0xFF:
		data = data[2:]
		bigEndian = true
	}
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	u16 := make([]uint16, len(data)/2)
	for i := range u16 {
		if bigEndian {
			u16[i] = binary.BigEndian.Uint16(data[i*2:])
		} else {
			u16[i] = binary.LittleEndian.Uint16(data[i*2:])
		}
	}
	return string(utf16.Decode(u16)), nil
}

func decodeUTF16LE(data []byte) string {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	u16 := make([]uint16, len(data)/2)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(data[i*2:])
	}
	return string(utf16.Decode(u16))
}

func encodeUTF16LEWithBOM(s string) []byte {
	body := stringToUTF16LE(s)
	out := make([]byte, 0, len(body)+2)
	out = append(out, 0xFF, 0xFE)
	out = append(out, body...)
	return out
}
