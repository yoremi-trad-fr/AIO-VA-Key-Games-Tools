package siglus

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
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

// DumpSSToTSV extrait les textes d'un fichier .ss et les sauvegarde en TSV
// Format TSV : index \t texte_original \t texte_traduit (vide)
func DumpSSToTSV(ssPath, tsvPath string) error {
	lines, err := DumpSS(ssPath)
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("index\toriginal\ttranslation\n")
	for _, l := range lines {
		if l.Text == "" {
			continue
		}
		// Encoder les sauts de ligne pour TSV
		text := strings.ReplaceAll(l.Text, "\n", "\\n")
		text = strings.ReplaceAll(text, "\t", "\\t")
		fmt.Fprintf(&sb, "%d\t%s\t\n", l.Index, text)
	}

	return os.WriteFile(tsvPath, []byte(sb.String()), 0644)
}

// InjectSS réinjecte les traductions depuis un TSV dans un fichier .ss
// Le TSV doit avoir le format : index \t original \t translation
func InjectSS(ssPath, tsvPath, outputPath string) error {
	// Charger la map des traductions
	tsvData, err := os.ReadFile(tsvPath)
	if err != nil {
		return fmt.Errorf("cannot read TSV: %w", err)
	}

	translations := make(map[int]string)
	for i, line := range strings.Split(string(tsvData), "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // skip header et lignes vides
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 || parts[2] == "" {
			continue
		}
		var idx int
		fmt.Sscanf(parts[0], "%d", &idx)
		text := parts[2]
		text = strings.ReplaceAll(text, "\\n", "\n")
		text = strings.ReplaceAll(text, "\\t", "\t")
		translations[idx] = text
	}

	if len(translations) == 0 {
		return fmt.Errorf("no translations found in TSV")
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
		tsvName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())) + ".tsv"
		tsvPath := filepath.Join(outputDir, tsvName)
		if err := DumpSSToTSV(ssPath, tsvPath); err != nil {
			fmt.Printf("[WARN] %s: %v\n", e.Name(), err)
			continue
		}
		count++
	}
	fmt.Printf("Dumped %d ss files to TSV in %s\n", count, outputDir)
	return nil
}

// InjectSSDir réinjecte tous les TSV d'un dossier dans les .ss correspondants
func InjectSSDir(ssDir, tsvDir, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(tsvDir)
	if err != nil {
		return err
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".tsv") {
			continue
		}
		baseName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		ssPath := filepath.Join(ssDir, baseName+".ss")
		tsvPath := filepath.Join(tsvDir, e.Name())
		outPath := filepath.Join(outputDir, baseName+".ss")

		if _, err := os.Stat(ssPath); os.IsNotExist(err) {
			fmt.Printf("[SKIP] %s (no matching .ss)\n", e.Name())
			continue
		}
		if err := InjectSS(ssPath, tsvPath, outPath); err != nil {
			fmt.Printf("[WARN] %s: %v\n", e.Name(), err)
			continue
		}
		count++
	}
	fmt.Printf("Injected %d files → %s\n", count, outputDir)
	return nil
}

// ─── helpers internes ───────────────────────────────────────

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
	names := make([]string, len(GameKeys))
	for i, gk := range GameKeys {
		names[i] = gk.Name
	}
	return names
}
