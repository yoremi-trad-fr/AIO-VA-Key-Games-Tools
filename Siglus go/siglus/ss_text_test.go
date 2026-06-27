package siglus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestFormatSSTextDumpMatchesSiglusToolsStyle(t *testing.T) {
	lines := []SSLine{
		{Index: 0, Text: "0"},
		{Index: 3, Text: "始まりは無機質な電子音だった。"},
		{Index: 4, Text: "It began with an unnatural beep."},
	}

	got := FormatSSTextDump(lines, SSDumpOptions{CopyText: true})
	want := "○0000000003○始まりは無機質な電子音だった。\r\n" +
		"●0000000003●始まりは無機質な電子音だった。\r\n\r\n"
	if got != want {
		t.Fatalf("unexpected dump:\n%q", got)
	}
}

func TestReadSSTranslationsSupportsSiglusToolsText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.ss.txt")
	text := "○0000000003○原文\r\n●0000000003●Traduction\r\n\r\n" +
		"○0000000007○削除\r\n●0000000007●\r\n\r\n"
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatalf("write text: %v", err)
	}

	got, err := ReadSSTranslations(path)
	if err != nil {
		t.Fatalf("read translations: %v", err)
	}
	if got[3] != "Traduction" {
		t.Fatalf("index 3 = %q", got[3])
	}
	if value, ok := got[7]; !ok || value != "" {
		t.Fatalf("index 7 = %q, ok=%v", value, ok)
	}
}

func TestReadSSTranslationsSupportsSiglusToolsXLSX(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.ss.xlsx")
	book := excelize.NewFile()
	defer func() { _ = book.Close() }()
	sheet := "sample.ss"
	if err := book.SetSheetName(book.GetSheetName(0), sheet); err != nil {
		t.Fatalf("rename sheet: %v", err)
	}
	lines := []SSLine{
		{Index: 3, Text: "原文"},
		{Index: 7, Text: "削除"},
	}
	if _, err := writeSSXLSXSheet(book, sheet, sheet, lines, SSDumpOptions{CopyText: true, ExportAllText: true}); err != nil {
		t.Fatalf("write sheet: %v", err)
	}
	if err := book.SetCellValue(sheet, "C3", ""); err != nil {
		t.Fatalf("clear translation: %v", err)
	}
	if err := book.SaveAs(path); err != nil {
		t.Fatalf("save xlsx: %v", err)
	}

	got, err := ReadSSTranslations(path)
	if err != nil {
		t.Fatalf("read translations: %v", err)
	}
	if got[3] != "原文" {
		t.Fatalf("index 3 = %q", got[3])
	}
	if value, ok := got[7]; !ok || value != "" {
		t.Fatalf("index 7 = %q, ok=%v", value, ok)
	}
}

func TestReadSSXLSXWorkbookUsesFullNameFromD1(t *testing.T) {
	path := filepath.Join(t.TempDir(), "single.xlsx")
	fullName := strings.Repeat("a", 32) + ".ss"
	book := excelize.NewFile()
	defer func() { _ = book.Close() }()
	used := map[string]bool{}
	sheet := ssExcelSheetName(fullName, used)
	if len([]rune(sheet)) != 31 {
		t.Fatalf("sheet length = %d, want 31", len([]rune(sheet)))
	}
	if err := book.SetSheetName(book.GetSheetName(0), sheet); err != nil {
		t.Fatalf("rename sheet: %v", err)
	}
	if _, err := writeSSXLSXSheet(book, sheet, fullName, []SSLine{{Index: 1, Text: "テスト"}}, SSDumpOptions{CopyText: true, ExportAllText: true}); err != nil {
		t.Fatalf("write sheet: %v", err)
	}
	if err := book.SaveAs(path); err != nil {
		t.Fatalf("save xlsx: %v", err)
	}

	sheets, err := readSSXLSXWorkbookTranslations(path)
	if err != nil {
		t.Fatalf("read workbook: %v", err)
	}
	if len(sheets) != 1 {
		t.Fatalf("sheets = %d", len(sheets))
	}
	if sheets[0].SSName != fullName {
		t.Fatalf("SSName = %q, want %q", sheets[0].SSName, fullName)
	}
}
