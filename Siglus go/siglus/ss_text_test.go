package siglus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestFormatSSTextDumpExportsAllTextByDefault(t *testing.T) {
	lines := []SSLine{
		{Index: 0, Text: "0"},
		{Index: 3, Text: "始まりは無機質な電子音だった。"},
		{Index: 4, Text: "It began with an unnatural beep."},
	}

	got := FormatSSTextDump(lines, SSDumpOptions{CopyText: true})
	want := "○0000000000○0\r\n" +
		"●0000000000●0\r\n\r\n" +
		"○0000000003○始まりは無機質な電子音だった。\r\n" +
		"●0000000003●始まりは無機質な電子音だった。\r\n\r\n" +
		"○0000000004○It began with an unnatural beep.\r\n" +
		"●0000000004●It began with an unnatural beep.\r\n\r\n"
	if got != want {
		t.Fatalf("unexpected dump:\n%q", got)
	}
}

func TestFormatSSTextDumpSingleLineStaysInjectable(t *testing.T) {
	lines := []SSLine{
		{Index: 3, Text: "始まりは無機質な電子音だった。"},
		{Index: 4, Text: "It began with an unnatural beep."},
	}

	got := FormatSSTextDump(lines, SSDumpOptions{CopyText: true, SingleLine: true})
	want := "●0000000003●始まりは無機質な電子音だった。\r\n" +
		"●0000000004●It began with an unnatural beep.\r\n"
	if got != want {
		t.Fatalf("unexpected single-line dump:\n%q", got)
	}

	path := filepath.Join(t.TempDir(), "sample.ss.txt")
	if err := os.WriteFile(path, []byte(got), 0644); err != nil {
		t.Fatalf("write text: %v", err)
	}
	translations, err := ReadSSTranslations(path)
	if err != nil {
		t.Fatalf("read translations: %v", err)
	}
	if translations[3] != lines[0].Text {
		t.Fatalf("index 3 = %q, want %q", translations[3], lines[0].Text)
	}
	if translations[4] != lines[1].Text {
		t.Fatalf("index 4 = %q, want %q", translations[4], lines[1].Text)
	}
}

func TestFormatSSTextDumpDialogueOnlyFiltersKnownTechnicalTags(t *testing.T) {
	lines := []SSLine{
		{Index: 1, Text: "CG"},
		{Index: 2, Text: "M"},
		{Index: 3, Text: "_system"},
		{Index: 4, Text: "bg_001"},
		{Index: 5, Text: "BGM01"},
		{Index: 6, Text: "se_click"},
		{Index: 7, Text: "fg01"},
		{Index: 8, Text: "ef-fade"},
		{Index: 9, Text: "si/title"},
		{Index: 10, Text: "tp:menu"},
		{Index: 11, Text: "md.001"},
		{Index: 12, Text: "ja"},
		{Index: 13, Text: "en"},
		{Index: 14, Text: "dummy"},
		{Index: 15, Text: "attack"},
		{Index: 16, Text: "intro1"},
		{Index: 17, Text: "sp04"},
		{Index: 18, Text: "sr00002"},
		{Index: 19, Text: "m02"},
		{Index: 20, Text: "$$fa_sway"},
		{Index: 21, Text: "town_smoke"},
		{Index: 22, Text: "Harmonia_trophy10"},
		{Index: 23, Text: "99_EF_add"},
		{Index: 24, Text: "It began with an unnatural beep."},
		{Index: 25, Text: "See you tomorrow."},
		{Index: 26, Text: "Sister"},
		{Index: 27, Text: "background"},
		{Index: 28, Text: "BGM is playing."},
		{Index: 29, Text: "P.S."},
		{Index: 30, Text: "始まりは無機質な電子音だった。"},
	}

	got := FormatSSTextDump(lines, SSDumpOptions{SingleLine: true, DialogueOnly: true})
	exported := make(map[string]bool)
	for _, line := range strings.Split(got, "\r\n") {
		parts := strings.SplitN(line, "●", 3)
		if len(parts) == 3 {
			exported[parts[2]] = true
		}
	}
	for _, removed := range []string{"CG", "_system", "bg_001", "BGM01", "se_click", "fg01", "ef-fade", "si/title", "tp:menu", "md.001", "ja", "en", "dummy", "attack", "intro1", "sp04", "sr00002", "m02", "$$fa_sway", "town_smoke", "Harmonia_trophy10", "99_EF_add"} {
		if exported[removed] {
			t.Errorf("technical tag %q was exported:\n%s", removed, got)
		}
	}
	for _, kept := range []string{"It began with an unnatural beep.", "See you tomorrow.", "Sister", "background", "BGM is playing.", "P.S.", "始まりは無機質な電子音だった。"} {
		if !exported[kept] {
			t.Errorf("dialogue %q was removed:\n%s", kept, got)
		}
	}
}

func TestIsSSTechnicalTagIsConservativeWithEnglish(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"CG", true},
		{"M", true},
		{"L", true},
		{"S", true},
		{"_system", true},
		{"bg_001", true},
		{"BGM01", true},
		{"seClick", true},
		{"fg/chara", true},
		{"ef-fade", true},
		{"si:title", true},
		{"tp.menu", true},
		{"md#001", true},
		{"ja", true},
		{"en", true},
		{"pg", true},
		{"dummy", true},
		{"attack", true},
		{"tipitipidorothy", true},
		{"intro1", true},
		{"sp04", true},
		{"sr00002", true},
		{"m02", true},
		{"$fa_sandstorm", true},
		{"$$fa_sway", true},
		{"town_smoke", true},
		{"siona_mv", true},
		{"sys_sa_info02", true},
		{"cgs_si010105_011", true},
		{"CGM_SZ10", true},
		{"Harmonia_trophy10", true},
		{"99_EF_add", true},
		{"See", false},
		{"Seriously", false},
		{"Sister", false},
		{"background", false},
		{"BGM is playing.", false},
		{"P.S.", false},
		{"Title", false},
		{"Gallery", false},
		{"Ending", false},
		{"CHAPTER", false},
		{"I_need your help.", false},
	}

	for _, tt := range tests {
		if got := isSSTechnicalTag(tt.text); got != tt.want {
			t.Errorf("isSSTechnicalTag(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestFormatSSTextDumpJapaneseOnlyKeepsLegacyNonASCIIFilter(t *testing.T) {
	lines := []SSLine{
		{Index: 1, Text: "English dialogue"},
		{Index: 2, Text: "日本語の台詞"},
	}

	got := FormatSSTextDump(lines, SSDumpOptions{SingleLine: true, JapaneseOnly: true})
	if strings.Contains(got, "English dialogue") || !strings.Contains(got, "日本語の台詞") {
		t.Fatalf("unexpected Japanese-only dump:\n%s", got)
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
