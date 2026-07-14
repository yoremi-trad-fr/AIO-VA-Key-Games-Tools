package siglus

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/xuri/excelize/v2"
)

// strEntry représente une entrée dans la string index d'un fichier .ss
type strEntry struct {
	charOffset int
	charSize   int
}

// ssHeader correspond à scrhead dans structs.h
// Chaque Entry = {offset int32, size int32}
type ssHeader struct {
	HeaderSize int32
	Bytecode   pairVal // 0x04
	StrIndex   pairVal // 0x0c
	StrTable   pairVal // 0x14
	Labels     pairVal // 0x1c
	Markers    pairVal // 0x24
	Unk3       pairVal // 0x2c
	Unk4       pairVal // 0x34
	Unk5       pairVal // 0x3c
	Unk6       pairVal // 0x44
	Unk7       pairVal // 0x4c
	Unk8       pairVal // 0x54
	Unk9       pairVal // 0x5c
	Unk10      pairVal // 0x64
	Unk11      pairVal // 0x6c
	Unk12      pairVal // 0x74
	Unk13      pairVal // 0x7c
}

const ssHeaderSize = 4 + 16*8 // 132 octets

// SSLine représente une ligne de texte extraite d'un .ss
type SSLine struct {
	Index int
	Text  string
}

type ssXLSXSheetTranslations struct {
	SSName       string
	Translations map[int]string
}

// SSDumpOptions matches the text export switches from Siglus Tools 0.61.
type SSDumpOptions struct {
	CopyText      bool
	ExportAllText bool
	FullWidthOnly bool
	SingleLine    bool
	DialogueOnly  bool
	JapaneseOnly  bool
}

// DumpSS extrait les chaînes de texte d'un fichier .ss
// Retourne la liste des lignes avec leur index
func DumpSS(ssPath string) ([]SSLine, error) {
	buf, err := os.ReadFile(ssPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read ss file: %w", err)
	}

	if len(buf) < ssHeaderSize {
		return nil, fmt.Errorf("file too small to be a valid .ss")
	}

	hdr := readSSHeader(buf)

	// Lire la string index
	idxOffset := int(hdr.StrIndex.Offset)
	idxCount := int(hdr.StrIndex.Size)
	if idxOffset+idxCount*8 > len(buf) {
		return nil, fmt.Errorf("string index out of bounds")
	}

	// Lire la string table
	tblOffset := int(hdr.StrTable.Offset)

	lines := make([]SSLine, 0, idxCount)

	for i := 0; i < idxCount; i++ {
		// Entry : {offset int32, size int32} en chars UTF-16
		entryOff := idxOffset + i*8
		strCharOffset := int(binary.LittleEndian.Uint32(buf[entryOff:]))
		strCharSize := int(binary.LittleEndian.Uint32(buf[entryOff+4:]))

		if strCharSize == 0 {
			lines = append(lines, SSLine{Index: i, Text: ""})
			continue
		}

		// Lire les chars UTF-16 depuis la string table
		byteOffset := tblOffset + strCharOffset*2
		byteSize := strCharSize * 2
		if byteOffset+byteSize > len(buf) {
			return nil, fmt.Errorf("string %d out of bounds", i)
		}

		u16 := make([]uint16, strCharSize)
		for j := range u16 {
			u16[j] = binary.LittleEndian.Uint16(buf[byteOffset+j*2:])
		}

		// Déchiffrement XOR : chaque string est XOR avec (index * 0x7087)
		key := uint16(i * 0x7087)
		for j := range u16 {
			u16[j] ^= key
		}

		text := string(utf16.Decode(u16))
		lines = append(lines, SSLine{Index: i, Text: text})
	}

	return lines, nil
}

// DumpSSToText extrait les textes d'un fichier .ss au format txt de Siglus Tools.
func DumpSSToText(ssPath, txtPath string, opts SSDumpOptions) error {
	lines, err := DumpSS(ssPath)
	if err != nil {
		return err
	}

	text := FormatSSTextDump(lines, opts)
	return os.WriteFile(txtPath, []byte(text), 0644)
}

// DumpSSToXLSX extrait les textes d'un fichier .ss au format Excel Siglus Tools.
func DumpSSToXLSX(ssPath, xlsxPath string, opts SSDumpOptions) error {
	lines, err := DumpSS(ssPath)
	if err != nil {
		return err
	}

	book := excelize.NewFile()
	defer func() { _ = book.Close() }()

	used := map[string]bool{}
	defaultSheet := book.GetSheetName(0)
	sheetName := ssExcelSheetName(filepath.Base(ssPath), used)
	if err := book.SetSheetName(defaultSheet, sheetName); err != nil {
		return err
	}
	if _, err := writeSSXLSXSheet(book, sheetName, filepath.Base(ssPath), lines, opts); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(xlsxPath), 0755); err != nil {
		return err
	}
	return book.SaveAs(xlsxPath)
}

// DumpSSToTSV is kept for older callers, but now writes the Siglus Tools text format.
func DumpSSToTSV(ssPath, tsvPath string) error {
	return DumpSSToText(ssPath, tsvPath, SSDumpOptions{})
}

func FormatSSTextDump(lines []SSLine, opts SSDumpOptions) string {
	var sb strings.Builder
	for _, l := range lines {
		if !shouldExportSSLine(l, opts) {
			continue
		}
		if opts.SingleLine {
			// Keep the translation marker so this compact dump can be edited and
			// passed directly to InjectSS without a conversion step.
			fmt.Fprintf(&sb, "●%010d●%s\r\n", l.Index, l.Text)
			continue
		}
		fmt.Fprintf(&sb, "○%010d○%s\r\n", l.Index, l.Text)
		if opts.CopyText {
			fmt.Fprintf(&sb, "●%010d●%s\r\n\r\n", l.Index, l.Text)
		} else {
			fmt.Fprintf(&sb, "●%010d●\r\n\r\n", l.Index)
		}
	}
	return sb.String()
}

// InjectSS réinjecte les traductions depuis un TSV dans un fichier .ss
// Le TSV doit avoir le format : index \t original \t translation
func InjectSS(ssPath, tsvPath, outputPath string) error {
	// Charger la map des traductions
	translations, err := ReadSSTranslations(tsvPath)
	if err != nil {
		return err
	}
	return injectSSWithTranslations(ssPath, translations, outputPath)
}

func injectSSWithTranslations(ssPath string, translations map[int]string, outputPath string) error {
	if len(translations) == 0 {
		return fmt.Errorf("no translations found")
	}

	// Lire le .ss original
	buf, err := os.ReadFile(ssPath)
	if err != nil {
		return fmt.Errorf("cannot read ss file: %w", err)
	}

	hdr := readSSHeader(buf)
	idxOffset := int(hdr.StrIndex.Offset)
	idxCount := int(hdr.StrIndex.Size)
	tblOffset := int(hdr.StrTable.Offset)
	tblSize := int(hdr.StrTable.Size) * 2 // en bytes

	entries := make([]strEntry, idxCount)
	for i := 0; i < idxCount; i++ {
		off := idxOffset + i*8
		entries[i].charOffset = int(binary.LittleEndian.Uint32(buf[off:]))
		entries[i].charSize = int(binary.LittleEndian.Uint32(buf[off+4:]))
	}

	// Reconstruire la string table avec les traductions
	// On construit une nouvelle table en remplaçant les strings traduites
	newTbl := make([]byte, tblSize)
	copy(newTbl, buf[tblOffset:tblOffset+tblSize])

	newEntries := make([]strEntry, idxCount)
	copy(newEntries, entries)

	// Pour les strings traduits, si même taille → patch in-place
	// Si taille différente → on doit reconstruire la table complète
	needRebuild := false
	for i, tr := range translations {
		if i >= idxCount {
			continue
		}
		newU16 := encryptString(tr, i)
		if len(newU16) != entries[i].charSize {
			needRebuild = true
			break
		}
	}

	if needRebuild {
		// Reconstruction complète de la string table
		var newTblBuf []byte
		currentOffset := 0
		for i := 0; i < idxCount; i++ {
			var u16 []uint16
			if tr, ok := translations[i]; ok {
				u16 = encryptString(tr, i)
			} else {
				// Lire l'original et le rechiffrer (il est déjà chiffré dans le buf)
				origOff := tblOffset + entries[i].charOffset*2
				origSize := entries[i].charSize
				u16 = make([]uint16, origSize)
				for j := range u16 {
					u16[j] = binary.LittleEndian.Uint16(buf[origOff+j*2:])
				}
			}
			newEntries[i].charOffset = currentOffset
			newEntries[i].charSize = len(u16)
			b := make([]byte, len(u16)*2)
			for j, v := range u16 {
				binary.LittleEndian.PutUint16(b[j*2:], v)
			}
			newTblBuf = append(newTblBuf, b...)
			currentOffset += len(u16)
		}

		// Reconstruire le fichier complet
		return rebuildSS(buf, hdr, newEntries, newTblBuf, outputPath)
	}

	// Patch in-place — tailles identiques
	for i, tr := range translations {
		if i >= idxCount {
			continue
		}
		u16 := encryptString(tr, i)
		byteOff := tblOffset + entries[i].charOffset*2
		for j, v := range u16 {
			binary.LittleEndian.PutUint16(buf[byteOff+j*2:], v)
		}
	}

	return os.WriteFile(outputPath, buf, 0644)
}

// DumpSSDir extrait tous les .ss d'un dossier vers des TSV
func DumpSSDir(inputDir, outputDir string) error {
	return DumpSSDirWithOptions(inputDir, outputDir, SSDumpOptions{})
}

// DumpSSDirWithOptions extrait tous les .ss d'un dossier vers des .txt Siglus Tools.
func DumpSSDirWithOptions(inputDir, outputDir string, opts SSDumpOptions) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".ss") {
			continue
		}
		ssPath := filepath.Join(inputDir, e.Name())
		txtName := e.Name() + ".txt"
		txtPath := filepath.Join(outputDir, txtName)
		if err := DumpSSToText(ssPath, txtPath, opts); err != nil {
			fmt.Printf("[WARN] %s: %v\n", e.Name(), err)
			continue
		}
		if info, err := os.Stat(txtPath); err == nil && info.Size() == 0 {
			_ = os.Remove(txtPath)
			continue
		}
		count++
	}
	fmt.Printf("Dumped %d ss files to text in %s\n", count, outputDir)
	return nil
}

// DumpSSDirToXLSX extrait tous les .ss d'un dossier vers un .xlsx par fichier.
func DumpSSDirToXLSX(inputDir, outputDir string, opts SSDumpOptions) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".ss") {
			continue
		}
		ssPath := filepath.Join(inputDir, e.Name())
		lines, err := DumpSS(ssPath)
		if err != nil {
			fmt.Printf("[WARN] %s: %v\n", e.Name(), err)
			continue
		}
		if countExportableSSLines(lines, opts) == 0 {
			continue
		}
		xlsxPath := filepath.Join(outputDir, e.Name()+".xlsx")
		if err := DumpSSToXLSX(ssPath, xlsxPath, opts); err != nil {
			fmt.Printf("[WARN] %s: %v\n", e.Name(), err)
			continue
		}
		count++
	}
	fmt.Printf("Dumped %d ss files to xlsx in %s\n", count, outputDir)
	return nil
}

// DumpSSDirToSingleXLSX extrait tous les .ss d'un dossier dans un seul classeur.
func DumpSSDirToSingleXLSX(inputDir, xlsxPath string, opts SSDumpOptions) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	book := excelize.NewFile()
	defer func() { _ = book.Close() }()

	usedSheets := map[string]bool{}
	defaultSheet := book.GetSheetName(0)
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".ss") {
			continue
		}
		ssPath := filepath.Join(inputDir, e.Name())
		lines, err := DumpSS(ssPath)
		if err != nil {
			fmt.Printf("[WARN] %s: %v\n", e.Name(), err)
			continue
		}
		if countExportableSSLines(lines, opts) == 0 {
			continue
		}

		sheetName := ssExcelSheetName(e.Name(), usedSheets)
		if count == 0 {
			if err := book.SetSheetName(defaultSheet, sheetName); err != nil {
				return err
			}
		} else if _, err := book.NewSheet(sheetName); err != nil {
			return err
		}
		if _, err := writeSSXLSXSheet(book, sheetName, e.Name(), lines, opts); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("no dumpable text found in %s", inputDir)
	}
	if err := os.MkdirAll(filepath.Dir(xlsxPath), 0755); err != nil {
		return err
	}
	if err := book.SaveAs(xlsxPath); err != nil {
		return err
	}
	fmt.Printf("Dumped %d ss files to xlsx %s\n", count, xlsxPath)
	return nil
}

// InjectSSDir réinjecte tous les TSV d'un dossier dans les .ss correspondants
func InjectSSDir(ssDir, tsvDir, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	if err := copySceneRebuildInputs(ssDir, outputDir); err != nil {
		return err
	}
	entries, err := os.ReadDir(tsvDir)
	if err != nil {
		return err
	}
	count := 0
	for _, e := range entries {
		lowerName := strings.ToLower(e.Name())
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(lowerName, ".xlsx") {
			n, err := InjectSSWorkbook(ssDir, filepath.Join(tsvDir, e.Name()), outputDir)
			if err != nil {
				fmt.Printf("[WARN] %s: %v\n", e.Name(), err)
				continue
			}
			count += n
			continue
		}
		if !strings.HasSuffix(lowerName, ".txt") && !strings.HasSuffix(lowerName, ".tsv") {
			continue
		}
		if injectSSDirTextFile(ssDir, filepath.Join(tsvDir, e.Name()), outputDir, e.Name()) {
			count++
		}
	}
	fmt.Printf("Injected %d files → %s\n", count, outputDir)
	return nil
}

func copySceneRebuildInputs(ssDir, outputDir string) error {
	entries, err := os.ReadDir(ssDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".ss" && ext != ".bin" {
			continue
		}
		src := filepath.Join(ssDir, entry.Name())
		dst := filepath.Join(outputDir, entry.Name())
		if samePath(src, dst) {
			continue
		}
		if err := copyFileBytes(src, dst); err != nil {
			return fmt.Errorf("cannot copy %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func samePath(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA == nil && errB == nil {
		return strings.EqualFold(filepath.Clean(absA), filepath.Clean(absB))
	}
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func copyFileBytes(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// InjectSSWorkbook importe un classeur Excel Siglus Tools dans un dossier de .ss.
func InjectSSWorkbook(ssDir, xlsxPath, outputDir string) (int, error) {
	sheets, err := readSSXLSXWorkbookTranslations(xlsxPath)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, sheet := range sheets {
		ssName := ssFileNameFromWorkbookName(sheet.SSName)
		ssPath := filepath.Join(ssDir, ssName)
		outPath := filepath.Join(outputDir, ssName)
		if _, err := os.Stat(ssPath); os.IsNotExist(err) {
			fmt.Printf("[SKIP] %s (no matching .ss)\n", sheet.SSName)
			continue
		}
		if err := injectSSWithTranslations(ssPath, sheet.Translations, outPath); err != nil {
			fmt.Printf("[WARN] %s: %v\n", sheet.SSName, err)
			continue
		}
		count++
	}
	return count, nil
}

// ─── helpers internes ───────────────────────────────────────

func shouldExportSSLine(line SSLine, opts SSDumpOptions) bool {
	if line.Text == "" {
		return false
	}
	if opts.ExportAllText {
		return true
	}
	if opts.DialogueOnly && isSSTechnicalTag(line.Text) {
		return false
	}
	if opts.FullWidthOnly {
		return shouldDumpSSText(line.Text, true)
	}
	if opts.JapaneseOnly {
		return shouldDumpSSText(line.Text, false)
	}
	// Safe default: keep every non-empty string. Filtering must always be an
	// explicit choice so English dialogue cannot disappear silently.
	return true
}

func isSSTechnicalTag(text string) bool {
	tag := strings.TrimSpace(text)
	if tag == "" {
		return false
	}

	switch strings.ToUpper(tag) {
	case "CG", "M", "L", "S":
		return true
	}
	switch tag {
	case "a", "b", "ja", "en", "pg", "dummy", "attack", "tipitipidorothy":
		return true
	}
	if strings.HasPrefix(tag, "_") {
		return true
	}
	if strings.IndexFunc(tag, unicode.IsSpace) >= 0 {
		return false
	}
	if strings.HasPrefix(tag, "$") {
		return true
	}
	if strings.ContainsRune(tag, '_') && isSSTechnicalIdentifier(tag) {
		return true
	}

	// Asset and control identifiers normally use a digit, separator or
	// camel-case suffix (bg_001, BGM01, seClick, ...). Requiring that shape
	// avoids treating ordinary English words such as "See" or "Sister" as
	// tags merely because they start with "se" or "si".
	lower := strings.ToLower(tag)
	for _, prefix := range []string{"intro", "bgm", "bg", "se", "fg", "ef", "si", "tp", "md", "sp", "sr", "m"} {
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		rest := tag[len(prefix):]
		if rest == "" {
			return true
		}
		first := rest[0]
		return first >= '0' && first <= '9' ||
			first >= 'A' && first <= 'Z' ||
			strings.ContainsRune("_-./\\:@#", rune(first))
	}
	return false
}

func isSSTechnicalIdentifier(text string) bool {
	for _, r := range text {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			continue
		}
		if strings.ContainsRune("_$-./\\:@#", r) {
			continue
		}
		return false
	}
	return true
}

func countExportableSSLines(lines []SSLine, opts SSDumpOptions) int {
	count := 0
	for _, line := range lines {
		if shouldExportSSLine(line, opts) {
			count++
		}
	}
	return count
}

func shouldDumpSSText(text string, fullWidthOnly bool) bool {
	if fullWidthOnly {
		for _, r := range text {
			if isSiglusHalfWidth(r) {
				return false
			}
		}
		return true
	}
	for _, r := range text {
		if !isSiglusHalfWidth(r) {
			return true
		}
	}
	return false
}

func isSiglusHalfWidth(r rune) bool {
	return r <= 0x7F
}

func ReadSSTranslations(path string) (map[int]string, error) {
	if strings.EqualFold(filepath.Ext(path), ".xlsx") {
		return readSSXLSXTranslations(path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read text dump: %w", err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if strings.Contains(text, "●") {
		return parseSSMarkerTranslations(text), nil
	}
	return parseSSTSVTranslations(text), nil
}

func parseSSMarkerTranslations(text string) map[int]string {
	translations := make(map[int]string)
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "●") {
			continue
		}
		rest := strings.TrimPrefix(line, "●")
		pos := strings.Index(rest, "●")
		if pos < 0 {
			continue
		}
		idx, err := strconv.Atoi(rest[:pos])
		if err != nil {
			continue
		}
		translations[idx] = rest[pos+len("●"):]
	}
	return translations
}

func parseSSTSVTranslations(text string) map[int]string {
	translations := make(map[int]string)
	for i, line := range strings.Split(text, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 || parts[2] == "" {
			continue
		}
		idx, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		value := strings.ReplaceAll(parts[2], "\\n", "\n")
		value = strings.ReplaceAll(value, "\\t", "\t")
		translations[idx] = value
	}
	return translations
}

func ssNameFromTextDump(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, ".ss.txt") {
		return name[:len(name)-len(".ss.txt")]
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func injectSSDirTextFile(ssDir, textPath, outputDir, textName string) bool {
	baseName := ssNameFromTextDump(textName)
	ssPath := filepath.Join(ssDir, baseName+".ss")
	outPath := filepath.Join(outputDir, baseName+".ss")

	if _, err := os.Stat(ssPath); os.IsNotExist(err) {
		fmt.Printf("[SKIP] %s (no matching .ss)\n", textName)
		return false
	}
	if err := InjectSS(ssPath, textPath, outPath); err != nil {
		fmt.Printf("[WARN] %s: %v\n", textName, err)
		return false
	}
	return true
}

func writeSSXLSXSheet(book *excelize.File, sheetName, fileName string, lines []SSLine, opts SSDumpOptions) (int, error) {
	if err := book.SetColWidth(sheetName, "A", "A", 8); err != nil {
		return 0, err
	}
	if err := book.SetColWidth(sheetName, "B", "C", 64); err != nil {
		return 0, err
	}
	headers := []any{"Index", "Text", "Translation"}
	if sheetName != fileName {
		headers = append(headers, fileName)
	}
	if err := book.SetSheetRow(sheetName, "A1", &headers); err != nil {
		return 0, err
	}

	row := 2
	count := 0
	for _, line := range lines {
		if !shouldExportSSLine(line, opts) {
			continue
		}
		translation := ""
		if opts.CopyText {
			translation = line.Text
		}
		values := []any{line.Index, line.Text, translation}
		cell := fmt.Sprintf("A%d", row)
		if err := book.SetSheetRow(sheetName, cell, &values); err != nil {
			return 0, err
		}
		row++
		count++
	}
	return count, nil
}

func readSSXLSXTranslations(path string) (map[int]string, error) {
	sheets, err := readSSXLSXWorkbookTranslations(path)
	if err != nil {
		return nil, err
	}
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no translation sheet found in %s", path)
	}
	return sheets[0].Translations, nil
}

func readSSXLSXWorkbookTranslations(path string) ([]ssXLSXSheetTranslations, error) {
	book, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read xlsx: %w", err)
	}
	defer func() { _ = book.Close() }()

	var sheets []ssXLSXSheetTranslations
	for _, sheetName := range book.GetSheetList() {
		translations := map[int]string{}
		rows, err := book.GetRows(sheetName)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if len(row) == 0 {
				continue
			}
			idx, err := parseSSXLSXIndex(row[0])
			if err != nil {
				continue
			}
			translation := ""
			if len(row) >= 3 {
				translation = row[2]
			}
			translations[idx] = translation
		}
		if len(translations) == 0 {
			continue
		}
		ssName := sheetName
		fullName, _ := book.GetCellValue(sheetName, "D1")
		if fullName != "" || len([]rune(sheetName)) >= 31 {
			if strings.TrimSpace(fullName) != "" {
				ssName = strings.TrimSpace(fullName)
			}
		}
		sheets = append(sheets, ssXLSXSheetTranslations{
			SSName:       ssName,
			Translations: translations,
		})
	}
	return sheets, nil
}

func parseSSXLSXIndex(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty index")
	}
	if idx, err := strconv.Atoi(value); err == nil {
		return idx, nil
	}
	if dot := strings.IndexByte(value, '.'); dot >= 0 {
		whole := value[:dot]
		fraction := strings.Trim(value[dot+1:], "0")
		if fraction == "" {
			return strconv.Atoi(whole)
		}
	}
	return 0, fmt.Errorf("invalid index: %s", value)
}

func ssFileNameFromWorkbookName(name string) string {
	name = strings.TrimSpace(name)
	if strings.HasSuffix(strings.ToLower(name), ".ss") {
		return name
	}
	return name + ".ss"
}

func ssExcelSheetName(fileName string, used map[string]bool) string {
	clean := sanitizeSSExcelSheetName(fileName)
	if clean == "" {
		clean = "Sheet"
	}
	base := truncateRunes(clean, 31)
	name := base
	for n := 1; used[strings.ToLower(name)]; n++ {
		suffix := fmt.Sprintf("_%d", n)
		name = truncateRunes(base, 31-len([]rune(suffix))) + suffix
	}
	used[strings.ToLower(name)] = true
	return name
}

func sanitizeSSExcelSheetName(name string) string {
	var sb strings.Builder
	for _, r := range name {
		switch r {
		case '[', ']', ':', '*', '?', '/', '\\':
			sb.WriteRune('_')
		default:
			if r < 0x20 {
				sb.WriteRune('_')
			} else {
				sb.WriteRune(r)
			}
		}
	}
	clean := strings.TrimSpace(sb.String())
	clean = strings.Trim(clean, "'")
	return clean
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func readSSHeader(buf []byte) ssHeader {
	var h ssHeader
	h.HeaderSize = int32(binary.LittleEndian.Uint32(buf[0:]))
	pairs := []*pairVal{
		&h.Bytecode, &h.StrIndex, &h.StrTable, &h.Labels, &h.Markers,
		&h.Unk3, &h.Unk4, &h.Unk5, &h.Unk6, &h.Unk7,
		&h.Unk8, &h.Unk9, &h.Unk10, &h.Unk11, &h.Unk12, &h.Unk13,
	}
	for i, p := range pairs {
		off := 4 + i*8
		p.Offset = int32(binary.LittleEndian.Uint32(buf[off:]))
		p.Size = int32(binary.LittleEndian.Uint32(buf[off+4:]))
	}
	return h
}

// encryptString chiffre une string Go pour l'écriture dans la string table .ss
func encryptString(s string, index int) []uint16 {
	u16 := utf16.Encode([]rune(s))
	key := uint16(index * 0x7087)
	for i := range u16 {
		u16[i] ^= key
	}
	return u16
}

// rebuildSS reconstruit un fichier .ss avec une nouvelle string table
func rebuildSS(orig []byte, hdr ssHeader, newEntries []strEntry, newTbl []byte, outputPath string) error {
	tblOffset := int(hdr.StrTable.Offset)
	idxOffset := int(hdr.StrIndex.Offset)
	idxCount := int(hdr.StrIndex.Size)

	oldTblSize := int(hdr.StrTable.Size) * 2
	newTblSize := len(newTbl)
	diff := newTblSize - oldTblSize

	// Construire le nouveau fichier
	// Partie avant la string table
	var out []byte
	out = append(out, orig[:tblOffset]...)
	out = append(out, newTbl...)
	// Partie après la string table
	afterOld := tblOffset + oldTblSize
	suffix := make([]byte, len(orig)-afterOld)
	copy(suffix, orig[afterOld:])

	// Ajuster les offsets dans le header si la taille a changé
	if diff != 0 {
		// Met à jour StrTable.Size dans le header de out
		newStrTableCharSize := newTblSize / 2
		binary.LittleEndian.PutUint32(out[4+2*8+4:], uint32(newStrTableCharSize))

		// Décaler tous les offsets après tblOffset dans le header
		adjustHeaderOffsets(out, tblOffset, diff)
	}

	// Mettre à jour la string index avec les nouveaux offsets/tailles
	for i := 0; i < idxCount; i++ {
		off := idxOffset + i*8
		binary.LittleEndian.PutUint32(out[off:], uint32(newEntries[i].charOffset))
		binary.LittleEndian.PutUint32(out[off+4:], uint32(newEntries[i].charSize))
	}

	out = append(out, suffix...)
	return os.WriteFile(outputPath, out, 0644)
}

// adjustHeaderOffsets décale tous les offsets du header ssHeader qui sont > threshold
func adjustHeaderOffsets(buf []byte, threshold, delta int) {
	for i := 0; i < 16; i++ {
		off := 4 + i*8
		v := int(binary.LittleEndian.Uint32(buf[off:]))
		if v > threshold {
			binary.LittleEndian.PutUint32(buf[off:], uint32(v+delta))
		}
	}
}

// GameNameList retourne les noms de jeux disponibles
func GameNameList() []string {
	names := make([]string, 0, len(GameKeys))
	seen := make(map[string]bool, len(GameKeys))
	for _, gk := range GameKeys {
		name := displayGameKeyName(gk.Name)
		if seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}
