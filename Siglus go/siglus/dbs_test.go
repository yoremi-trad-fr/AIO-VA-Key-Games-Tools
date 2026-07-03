package siglus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestDBSUnicodePackUnpack(t *testing.T) {
	dir := t.TempDir()
	rawPath := filepath.Join(dir, "sample.dbs.out")
	txtPath := filepath.Join(dir, "sample.dbs.txt")
	dbsPath := filepath.Join(dir, "sample.dbs")
	unpackedRaw := filepath.Join(dir, "unpacked.dbs.out")
	unpackedTxt := filepath.Join(dir, "unpacked.dbs.txt")

	table := makeTestDBSTable(true)
	raw, err := buildDBSRaw(table, "shift-jis")
	if err != nil {
		t.Fatalf("build raw: %v", err)
	}
	if err := os.WriteFile(rawPath, raw, 0644); err != nil {
		t.Fatalf("write raw: %v", err)
	}

	txt := "Unicode\n" +
		"[0001]\n" +
		"○01○こんにちは\n" +
		"●01●Bonjour\n\n" +
		"{02}7\n" +
		"<02>42\n\n\n" +
		"[0002]\n" +
		"○01○\n" +
		"●01●Deuxieme ligne\n\n" +
		"{02}11\n" +
		"<02>99\n\n\n"
	if err := os.WriteFile(txtPath, encodeUTF16LEWithBOM(txt), 0644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	if err := PackDBS(rawPath, txtPath, dbsPath, 17, false); err != nil {
		t.Fatalf("pack DBS: %v", err)
	}
	if err := UnpackDBS(dbsPath, unpackedRaw, unpackedTxt, true); err != nil {
		t.Fatalf("unpack DBS: %v", err)
	}

	unpacked, err := os.ReadFile(unpackedRaw)
	if err != nil {
		t.Fatalf("read unpacked raw: %v", err)
	}
	parsed, err := parseDBSRaw(unpacked, true, "shift-jis")
	if err != nil {
		t.Fatalf("parse unpacked raw: %v", err)
	}
	if got := parsed.rows[0][0].text; got != "Bonjour" {
		t.Fatalf("row 0 text = %q", got)
	}
	if got := parsed.rows[0][1].num; got != 42 {
		t.Fatalf("row 0 num = %d", got)
	}
	if got := parsed.rows[1][0].text; got != "Deuxieme ligne" {
		t.Fatalf("row 1 text = %q", got)
	}
	if got := parsed.rows[1][1].num; got != 99 {
		t.Fatalf("row 1 num = %d", got)
	}

	txtBytes, err := os.ReadFile(unpackedTxt)
	if err != nil {
		t.Fatalf("read unpacked txt: %v", err)
	}
	txtOut, err := decodeUTF16Text(txtBytes)
	if err != nil {
		t.Fatalf("decode unpacked txt: %v", err)
	}
	if !strings.Contains(txtOut, "●01●Bonjour") {
		t.Fatalf("unpacked text does not contain translation:\n%s", txtOut)
	}
}

func TestDBSASCIIPackUnpackShiftJIS(t *testing.T) {
	dir := t.TempDir()
	rawPath := filepath.Join(dir, "sample.dbs.out")
	txtPath := filepath.Join(dir, "sample.dbs.txt")
	dbsPath := filepath.Join(dir, "sample.dbs")
	unpackedRaw := filepath.Join(dir, "unpacked.dbs.out")

	table := makeTestDBSTable(false)
	table.rows[0][0].text = "かな"
	table.rows[1][0].text = "テスト"
	raw, err := buildDBSRaw(table, "shift-jis")
	if err != nil {
		t.Fatalf("build raw: %v", err)
	}
	if err := os.WriteFile(rawPath, raw, 0644); err != nil {
		t.Fatalf("write raw: %v", err)
	}

	txt := "ASCII\n" +
		"[0001]\n" +
		"○01○かな\n" +
		"●01●カナ\n\n" +
		"{02}7\n" +
		"<02>8\n\n\n" +
		"[0002]\n" +
		"○01○テスト\n" +
		"●01●テキスト\n\n" +
		"{02}11\n" +
		"<02>12\n\n\n"
	if err := os.WriteFile(txtPath, encodeUTF16LEWithBOM(txt), 0644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	if err := PackDBSWithOptions(rawPath, txtPath, dbsPath, DBSPackOptions{
		CompressionLevel: 17,
		ANSIEncoding:     "shift-jis",
	}); err != nil {
		t.Fatalf("pack DBS: %v", err)
	}
	if err := UnpackDBS(dbsPath, unpackedRaw, filepath.Join(dir, "unpacked.dbs.txt"), true); err != nil {
		t.Fatalf("unpack DBS: %v", err)
	}

	unpacked, err := os.ReadFile(unpackedRaw)
	if err != nil {
		t.Fatalf("read unpacked raw: %v", err)
	}
	parsed, err := parseDBSRaw(unpacked, false, "shift-jis")
	if err != nil {
		t.Fatalf("parse unpacked raw: %v", err)
	}
	if got := parsed.rows[0][0].text; got != "カナ" {
		t.Fatalf("row 0 text = %q", got)
	}
	if got := parsed.rows[1][0].text; got != "テキスト" {
		t.Fatalf("row 1 text = %q", got)
	}
}

func TestDBSASCIIPackKeepsUnmodifiedGBKRows(t *testing.T) {
	dir := t.TempDir()
	rawPath := filepath.Join(dir, "sample.dbs.out")
	txtPath := filepath.Join(dir, "sample.dbs.txt")
	dbsPath := filepath.Join(dir, "sample.dbs")
	unpackedRaw := filepath.Join(dir, "unpacked.dbs.out")

	table := makeTestDBSTable(false)
	table.rows[0][0].text = "中文"
	table.rows[1][0].text = "保留"
	raw, err := buildDBSRaw(table, "gbk")
	if err != nil {
		t.Fatalf("build raw: %v", err)
	}
	if err := os.WriteFile(rawPath, raw, 0644); err != nil {
		t.Fatalf("write raw: %v", err)
	}

	txt := "ASCII\n" +
		"[0001]\n" +
		"●01●修改\n\n"
	if err := os.WriteFile(txtPath, encodeUTF16LEWithBOM(txt), 0644); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	if err := PackDBSWithOptions(rawPath, txtPath, dbsPath, DBSPackOptions{
		CompressionLevel: 17,
		ANSIEncoding:     "gbk",
	}); err != nil {
		t.Fatalf("pack DBS: %v", err)
	}
	if err := UnpackDBSWithEncoding(dbsPath, unpackedRaw, filepath.Join(dir, "unpacked.dbs.txt"), true, "gbk"); err != nil {
		t.Fatalf("unpack DBS: %v", err)
	}

	unpacked, err := os.ReadFile(unpackedRaw)
	if err != nil {
		t.Fatalf("read unpacked raw: %v", err)
	}
	parsed, err := parseDBSRaw(unpacked, false, "gbk")
	if err != nil {
		t.Fatalf("parse unpacked raw: %v", err)
	}
	if got := parsed.rows[0][0].text; got != "修改" {
		t.Fatalf("row 0 text = %q", got)
	}
	if got := parsed.rows[1][0].text; got != "保留" {
		t.Fatalf("row 1 text = %q", got)
	}
}

func TestDBSXLSXDumpAndBuild(t *testing.T) {
	dir := t.TempDir()
	sourceDBS := filepath.Join(dir, "source.dbs")
	xlsxPath := filepath.Join(dir, "source.dbs.xlsx")
	builtDBS := filepath.Join(dir, "built.dbs")
	unpackedRaw := filepath.Join(dir, "built.dbs.out")

	table := makeTestDBSTable(true)
	table.rows[0][0].text = "最初"
	table.rows[0][1].num = 7
	table.rows[1][0].text = "二番目"
	table.rows[1][1].num = 11
	if err := writePackedDBS(table, sourceDBS, DBSPackOptions{CompressionLevel: 17}); err != nil {
		t.Fatalf("write source dbs: %v", err)
	}

	if err := DumpDBSToXLSX(sourceDBS, xlsxPath, "shift-jis"); err != nil {
		t.Fatalf("dump xlsx: %v", err)
	}

	book, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	if got, _ := book.GetCellValue("Translation", "A1"); got != "#DATANO" {
		t.Fatalf("A1 = %q", got)
	}
	if got, _ := book.GetCellValue("Translation", "B2"); got != "S" {
		t.Fatalf("B2 = %q", got)
	}
	if err := book.SetCellValue("Translation", "B3", "Modifie"); err != nil {
		t.Fatalf("edit text: %v", err)
	}
	if err := book.SetCellValue("Translation", "C3", 42); err != nil {
		t.Fatalf("edit int: %v", err)
	}
	if err := book.Save(); err != nil {
		t.Fatalf("save xlsx: %v", err)
	}
	_ = book.Close()

	if err := BuildDBSFromXLSX(xlsxPath, builtDBS, DBSBuildOptions{
		CompressionLevel: 17,
		Unicode:          true,
	}); err != nil {
		t.Fatalf("build dbs: %v", err)
	}
	if err := UnpackDBS(builtDBS, unpackedRaw, filepath.Join(dir, "built.dbs.txt"), true); err != nil {
		t.Fatalf("unpack built dbs: %v", err)
	}
	raw, err := os.ReadFile(unpackedRaw)
	if err != nil {
		t.Fatalf("read raw: %v", err)
	}
	parsed, err := parseDBSRaw(raw, true, "shift-jis")
	if err != nil {
		t.Fatalf("parse raw: %v", err)
	}
	if got := parsed.rows[0][0].text; got != "Modifie" {
		t.Fatalf("row 0 text = %q", got)
	}
	if got := parsed.rows[0][1].num; got != 42 {
		t.Fatalf("row 0 num = %d", got)
	}
	if got := parsed.rows[1][0].text; got != "二番目" {
		t.Fatalf("row 1 text = %q", got)
	}
}

func TestDBSBuildDirFromXLSXASCII(t *testing.T) {
	dir := t.TempDir()
	xlsxDir := filepath.Join(dir, "xlsx")
	outDir := filepath.Join(dir, "dbs")
	if err := os.MkdirAll(xlsxDir, 0755); err != nil {
		t.Fatalf("mkdir xlsx: %v", err)
	}
	xlsxPath := filepath.Join(xlsxDir, "sample.xlsx")

	book := excelize.NewFile()
	sheet := book.GetSheetName(0)
	if err := book.SetSheetRow(sheet, "A1", &[]any{"#DATANO", 1, 2}); err != nil {
		t.Fatalf("row 1: %v", err)
	}
	if err := book.SetSheetRow(sheet, "A2", &[]any{"#DATATYPE", "S", "V"}); err != nil {
		t.Fatalf("row 2: %v", err)
	}
	if err := book.SetSheetRow(sheet, "A3", &[]any{10, "かな", 123}); err != nil {
		t.Fatalf("row 3: %v", err)
	}
	if err := book.SaveAs(xlsxPath); err != nil {
		t.Fatalf("save xlsx: %v", err)
	}
	_ = book.Close()

	if err := BuildDBSDirFromXLSX(xlsxDir, outDir, DBSBuildOptions{
		CompressionLevel: 17,
		Unicode:          false,
		ANSIEncoding:     "shift-jis",
	}); err != nil {
		t.Fatalf("build dir: %v", err)
	}
	outDBS := filepath.Join(outDir, "sample.dbs")
	rawPath := filepath.Join(dir, "sample.dbs.out")
	if err := UnpackDBSWithEncoding(outDBS, rawPath, filepath.Join(dir, "sample.dbs.txt"), true, "shift-jis"); err != nil {
		t.Fatalf("unpack ascii dbs: %v", err)
	}
	raw, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("read raw: %v", err)
	}
	parsed, err := parseDBSRaw(raw, false, "shift-jis")
	if err != nil {
		t.Fatalf("parse raw: %v", err)
	}
	if got := parsed.rows[0][0].text; got != "かな" {
		t.Fatalf("text = %q", got)
	}
	if got := parsed.rows[0][1].num; got != 123 {
		t.Fatalf("num = %d", got)
	}
}

func makeTestDBSTable(isUnicode bool) *dbsTable {
	headerData := make([]byte, 24)
	lineCount := uint32(2)
	dataCount := uint32(2)
	lineIndexOffset := uint32(28)
	dataFormatOffset := lineIndexOffset + lineCount*4
	lineDataIndexOffset := dataFormatOffset + dataCount*8
	textOffset := lineDataIndexOffset + lineCount*dataCount*4
	put := func(pos int, v uint32) {
		headerData[pos] = byte(v)
		headerData[pos+1] = byte(v >> 8)
		headerData[pos+2] = byte(v >> 16)
		headerData[pos+3] = byte(v >> 24)
	}
	put(0, lineCount)
	put(4, dataCount)
	put(8, lineIndexOffset)
	put(12, dataFormatOffset)
	put(16, lineDataIndexOffset)
	put(20, textOffset)

	return &dbsTable{
		header: dbsHeader{
			headerData:          headerData,
			lineCount:           lineCount,
			dataCount:           dataCount,
			lineIndexOffset:     lineIndexOffset,
			dataFormatOffset:    dataFormatOffset,
			lineDataIndexOffset: lineDataIndexOffset,
			textOffset:          textOffset,
		},
		isUnicode: isUnicode,
		lineIndex: []int32{1, 2},
		columns: []dbsColumn{
			{index: 1, typ: dbsTypeString},
			{index: 2, typ: dbsTypeInt},
		},
		rows: [][]dbsValue{
			{{text: "こんにちは"}, {num: 7}},
			{{text: ""}, {num: 11}},
		},
	}
}
