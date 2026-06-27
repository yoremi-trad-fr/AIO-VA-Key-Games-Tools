package siglus

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// DBSBuildOptions controls DBS creation from Siglus Tools-compatible XLSX files.
type DBSBuildOptions struct {
	CompressionLevel int
	FakeCompression  bool
	Unicode          bool
	ANSIEncoding     string
}

func DumpDBSToXLSX(inputPath, xlsxOutputPath, ansiEncoding string) error {
	table, _, err := readPackedDBS(inputPath, ansiEncoding)
	if err != nil {
		return err
	}

	book := excelize.NewFile()
	defer func() { _ = book.Close() }()

	defaultSheet := book.GetSheetName(0)
	if err := book.SetSheetName(defaultSheet, "Translation"); err != nil {
		return err
	}
	if err := writeDBSXLSXSheet(book, "Translation", table); err != nil {
		return err
	}
	if _, err := book.NewSheet("Text"); err != nil {
		return err
	}
	if err := writeDBSXLSXSheet(book, "Text", table); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(xlsxOutputPath), 0755); err != nil {
		return err
	}
	return book.SaveAs(xlsxOutputPath)
}

func BuildDBSFromXLSX(xlsxPath, outputPath string, opts DBSBuildOptions) error {
	table, err := readDBSTableFromXLSX(xlsxPath, opts)
	if err != nil {
		return err
	}
	return writePackedDBS(table, outputPath, DBSPackOptions{
		CompressionLevel: opts.CompressionLevel,
		FakeCompression:  opts.FakeCompression,
		ANSIEncoding:     opts.ANSIEncoding,
	})
}

func BuildDBSDirFromXLSX(inputDir, outputDir string, opts DBSBuildOptions) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".xlsx") {
			continue
		}
		inputPath := filepath.Join(inputDir, entry.Name())
		outputName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) + ".dbs"
		outputName = strings.ReplaceAll(outputName, ".dbs.dbs", ".dbs")
		outputPath := filepath.Join(outputDir, outputName)
		if err := BuildDBSFromXLSX(inputPath, outputPath, opts); err != nil {
			fmt.Printf("[WARN] %s: %v\n", entry.Name(), err)
			continue
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("no xlsx files built from %s", inputDir)
	}
	fmt.Printf("Built %d dbs files -> %s\n", count, outputDir)
	return nil
}

func readPackedDBS(inputPath, ansiEncoding string) (*dbsTable, []byte, error) {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read DBS: %w", err)
	}
	if len(buf) < 12 {
		return nil, nil, fmt.Errorf("file too small to be a DBS")
	}

	isUnicode := binary.LittleEndian.Uint32(buf[0:4]) != 0
	data := append([]byte(nil), buf[4:]...)
	xorDBS(data)

	decompressed := Decompress(data)
	if decompressed == nil {
		return nil, nil, fmt.Errorf("DBS decompression failed")
	}
	xorDBSLayer(decompressed)

	table, err := parseDBSRaw(decompressed, isUnicode, ansiEncoding)
	if err != nil {
		return nil, nil, err
	}
	return table, decompressed, nil
}

func writePackedDBS(table *dbsTable, outputPath string, opts DBSPackOptions) error {
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
	if table.isUnicode {
		binary.LittleEndian.PutUint32(out[0:4], 1)
	}
	out = append(out, packed...)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		return fmt.Errorf("cannot write DBS: %w", err)
	}
	return nil
}

func writeDBSXLSXSheet(book *excelize.File, sheet string, table *dbsTable) error {
	if err := book.SetColWidth(sheet, "A", "A", 10); err != nil {
		return err
	}
	headerIndex := []any{"#DATANO"}
	headerType := []any{"#DATATYPE"}
	for _, column := range table.columns {
		headerIndex = append(headerIndex, int(column.index))
		if column.typ == dbsTypeString {
			headerType = append(headerType, "S")
		} else {
			headerType = append(headerType, "V")
		}
	}
	if err := book.SetSheetRow(sheet, "A1", &headerIndex); err != nil {
		return err
	}
	if err := book.SetSheetRow(sheet, "A2", &headerType); err != nil {
		return err
	}
	for rowIdx, row := range table.rows {
		values := []any{int(table.lineIndex[rowIdx])}
		for colIdx, column := range table.columns {
			value := row[colIdx]
			if column.typ == dbsTypeString {
				values = append(values, value.text)
			} else {
				values = append(values, value.num)
			}
		}
		cell := fmt.Sprintf("A%d", rowIdx+3)
		if err := book.SetSheetRow(sheet, cell, &values); err != nil {
			return err
		}
	}
	return nil
}

func readDBSTableFromXLSX(path string, opts DBSBuildOptions) (*dbsTable, error) {
	book, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read xlsx: %w", err)
	}
	defer func() { _ = book.Close() }()

	sheet := firstDBSSheet(book)
	rows, err := book.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	dataIndexRow, dataTypeRow := -1, -1
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}
		switch strings.TrimSpace(row[0]) {
		case "#DATANO":
			dataIndexRow = i
		case "#DATATYPE":
			dataTypeRow = i
		}
	}
	if dataIndexRow < 0 || dataTypeRow < 0 {
		return nil, fmt.Errorf("data format error: #DATANO/#DATATYPE rows missing")
	}

	indexCells := rows[dataIndexRow]
	typeCells := rows[dataTypeRow]
	var columns []dbsColumn
	var sourceCols []int
	for col := 1; col < len(indexCells); col++ {
		index, err := parseDBSXLSXUint(indexCells[col])
		if err != nil {
			continue
		}
		typeText := ""
		if col < len(typeCells) {
			typeText = strings.TrimSpace(typeCells[col])
		}
		switch typeText {
		case "S":
			columns = append(columns, dbsColumn{index: index, typ: dbsTypeString})
		case "V":
			columns = append(columns, dbsColumn{index: index, typ: dbsTypeInt})
		default:
			continue
		}
		sourceCols = append(sourceCols, col)
	}
	if len(columns) == 0 {
		return nil, fmt.Errorf("data format error: no DBS columns found")
	}

	var lineIndex []int32
	var tableRows [][]dbsValue
	for rowNum := dataTypeRow + 1; rowNum < len(rows); rowNum++ {
		row := rows[rowNum]
		if len(row) == 0 {
			continue
		}
		index, err := parseDBSXLSXInt(row[0])
		if err != nil {
			continue
		}
		lineIndex = append(lineIndex, int32(index))
		values := make([]dbsValue, len(columns))
		for i, column := range columns {
			value := ""
			sourceCol := sourceCols[i]
			if sourceCol < len(row) {
				value = row[sourceCol]
			}
			if column.typ == dbsTypeString {
				values[i].text = value
			} else if parsed, err := parseUint32Text(value); err == nil {
				values[i].num = parsed
			}
		}
		tableRows = append(tableRows, values)
	}
	if len(lineIndex) == 0 {
		return nil, fmt.Errorf("data format error: no DBS rows found")
	}

	headerData := make([]byte, 24)
	lineCount := uint32(len(lineIndex))
	dataCount := uint32(len(columns))
	lineIndexOffset := uint32(0x1C)
	dataFormatOffset := lineIndexOffset + lineCount*4
	lineDataIndexOffset := dataFormatOffset + dataCount*8
	textOffset := lineDataIndexOffset + lineCount*dataCount*4
	binary.LittleEndian.PutUint32(headerData[0:4], lineCount)
	binary.LittleEndian.PutUint32(headerData[4:8], dataCount)
	binary.LittleEndian.PutUint32(headerData[8:12], lineIndexOffset)
	binary.LittleEndian.PutUint32(headerData[12:16], dataFormatOffset)
	binary.LittleEndian.PutUint32(headerData[16:20], lineDataIndexOffset)
	binary.LittleEndian.PutUint32(headerData[20:24], textOffset)

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
		isUnicode: opts.Unicode,
		lineIndex: lineIndex,
		columns:   columns,
		rows:      tableRows,
	}, nil
}

func firstDBSSheet(book *excelize.File) string {
	if sheet := book.GetSheetName(book.GetActiveSheetIndex()); sheet != "" {
		return sheet
	}
	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return ""
	}
	return sheets[0]
}

func parseDBSXLSXUint(value string) (uint32, error) {
	parsed, err := parseDBSXLSXInt(value)
	return uint32(parsed), err
}

func parseDBSXLSXInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed, nil
	}
	if dot := strings.IndexByte(value, '.'); dot >= 0 {
		whole := value[:dot]
		fraction := strings.Trim(value[dot+1:], "0")
		if fraction == "" {
			return strconv.Atoi(whole)
		}
	}
	return 0, fmt.Errorf("invalid integer: %s", value)
}
