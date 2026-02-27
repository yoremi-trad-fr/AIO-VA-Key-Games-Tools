package siglus

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"
)

// pairVal correspond à PAIRVAL du C : {offset int32, size int32}
type pairVal struct {
	Offset int32
	Size   int32
}

// pckHeader correspond exactement à PCKHDR du C
// Taille fixe : 4 + 10*8 + 4 + 4 = 92 octets
type pckHeader struct {
	HdrLen     int32
	Table1     pairVal
	GlobalVar  pairVal
	GlobalVarStr pairVal
	Name1      pairVal
	Name2      pairVal
	Name3      pairVal
	Name4      pairVal
	FnameStr   pairVal
	FileToc    pairVal
	Data       pairVal
	Encrypt2   int32
	Wtf        int32
}

// sectionNames correspond à la liste sections[] du C
var sectionNames = []string{"table1", "gvar", "gvarstr", "name1", "name2", "name3", "name4", "fname"}

// decrypt applique le déchiffrement XOR sur les données.
// Première passe : XOR avec key2 (universelle, 256 octets).
// Deuxième passe si encrypt2==1 : XOR avec key1 (spécifique au jeu, 16 octets).
func decrypt(data []byte, encrypt2 int32, key1 [16]byte) {
	for j := range data {
		data[j] ^= key2[j%256]
	}
	if encrypt2 == 1 {
		for j := range data {
			data[j] ^= key1[j%16]
		}
	}
}

// utf16LEToString convertit une suite de bytes UTF-16 LE en string Go (UTF-8)
func utf16LEToString(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u16 := make([]uint16, len(b)/2)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	// Trouver le null-terminator
	for i, v := range u16 {
		if v == 0 {
			u16 = u16[:i]
			break
		}
	}
	return string(utf16.Decode(u16))
}

// stringToUTF16LE convertit une string Go en bytes UTF-16 LE
func stringToUTF16LE(s string) []byte {
	u16 := utf16.Encode([]rune(s))
	b := make([]byte, len(u16)*2)
	for i, v := range u16 {
		binary.LittleEndian.PutUint16(b[i*2:], v)
	}
	return b
}

// ExtractPCK extrait tous les fichiers .ss depuis un Scene.pck.
// key1 est la clé spécifique au jeu (16 octets).
// outputDir est le dossier de sortie pour les .ss et les tables binaires.
func ExtractPCK(pckPath string, key1 [16]byte, outputDir string) error {
	// Lecture du fichier PCK complet
	buf, err := os.ReadFile(pckPath)
	if err != nil {
		return fmt.Errorf("cannot read PCK: %w", err)
	}

	// Lecture du header
	if len(buf) < 92 {
		return fmt.Errorf("file too small to be a valid PCK")
	}
	hdr := &pckHeader{}
	if err := readHeader(buf, hdr); err != nil {
		return err
	}

	expectedHdrLen := int32(92)
	if hdr.HdrLen != expectedHdrLen {
		return fmt.Errorf("wrong PCK header size: got %d, expected %d", hdr.HdrLen, expectedHdrLen)
	}

	// Créer le dossier de sortie
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output dir: %w", err)
	}

	// Extraire les tables binaires (table1, gvar, etc.)
	sections := getSectionPairs(hdr)
	for i, name := range sectionNames {
		if err := dumpSection(outputDir, name, buf, sections[i], sections[i+1]); err != nil {
			fmt.Printf("[WARN] cannot dump section %s: %v\n", name, err)
		}
	}

	// Lire la table des fichiers et la table des noms
	numFiles := int(hdr.FileToc.Size)
	tocOffset := int(hdr.FileToc.Offset)
	nameIdxOffset := int(hdr.Name4.Offset)
	fnameStrOffset := int(hdr.FnameStr.Offset)
	dataOffset := int(hdr.Data.Offset)

	fmt.Printf("Extracting %d files from %s...\n", numFiles, filepath.Base(pckPath))

	for i := 0; i < numFiles; i++ {
		// Lire le toc entry : {offset int32, size int32}
		tocEntry := tocOffset + i*8
		if tocEntry+8 > len(buf) {
			return fmt.Errorf("TOC entry %d out of bounds", i)
		}
		fileOffset := int(binary.LittleEndian.Uint32(buf[tocEntry:]))
		fileSize := int(binary.LittleEndian.Uint32(buf[tocEntry+4:]))

		// Lire le nom du fichier (UTF-16 LE)
		nameIdx := nameIdxOffset + i*8
		if nameIdx+8 > len(buf) {
			return fmt.Errorf("name index %d out of bounds", i)
		}
		nameCharOffset := int(binary.LittleEndian.Uint32(buf[nameIdx:]))
		nameCharSize := int(binary.LittleEndian.Uint32(buf[nameIdx+4:]))

		nameByteOffset := fnameStrOffset + nameCharOffset*2
		nameByteSize := nameCharSize * 2
		if nameByteOffset+nameByteSize > len(buf) {
			return fmt.Errorf("filename data out of bounds for file %d", i)
		}
		fname := utf16LEToString(buf[nameByteOffset : nameByteOffset+nameByteSize])
		outName := fname + ".ss"

		fmt.Printf("  [%d/%d] %s (offset=0x%X, size=%d)\n", i+1, numFiles, outName, fileOffset, fileSize)

		// Extraire et déchiffrer les données
		dataStart := dataOffset + fileOffset
		if dataStart+fileSize > len(buf) {
			return fmt.Errorf("data for %s out of bounds", fname)
		}

		// Copie locale pour ne pas modifier le buffer original
		fileData := make([]byte, fileSize)
		copy(fileData, buf[dataStart:dataStart+fileSize])

		// Déchiffrement
		decrypt(fileData, hdr.Encrypt2, key1)

		compLen := int(binary.LittleEndian.Uint32(fileData[0:4]))
		if compLen != fileSize {
			fmt.Printf("    [WARN] size mismatch: header=%d / toc=%d\n", compLen, fileSize)
		}

		// Décompression
		decompData := Decompress(fileData)
		if decompData == nil {
			return fmt.Errorf("decompression failed for %s", fname)
		}

		// Écriture
		outPath := filepath.Join(outputDir, outName)
		if err := os.WriteFile(outPath, decompData, 0644); err != nil {
			return fmt.Errorf("cannot write %s: %w", outName, err)
		}
	}

	fmt.Printf("Extraction complete → %s\n", outputDir)
	return nil
}

// RebuildPCK recompile un Scene.pck depuis un dossier de fichiers .ss.
// Utilise les tables binaires (*.bin) extraites lors de l'extraction.
// key1 est la clé spécifique au jeu.
func RebuildPCK(inputDir string, key1 [16]byte, wtfVal int32, outputPath string) error {
	// Lire les tables binaires
	sections := make([][]byte, len(sectionNames))
	for i, name := range sectionNames {
		data, err := os.ReadFile(filepath.Join(inputDir, name+".bin"))
		if err != nil {
			return fmt.Errorf("cannot read %s.bin: %w", name, err)
		}
		sections[i] = data
	}

	// Lire fname.bin et name4.bin pour la liste des fichiers
	fnameBin, err := os.ReadFile(filepath.Join(inputDir, "fname.bin"))
	if err != nil {
		return fmt.Errorf("cannot read fname.bin: %w", err)
	}
	name4Bin, err := os.ReadFile(filepath.Join(inputDir, "name4.bin"))
	if err != nil {
		return fmt.Errorf("cannot read name4.bin: %w", err)
	}

	// fname.bin : [int32 decSize][UTF-16 LE string data...]
	// name4.bin : [int32 count][{offset int32, size int32} * count]
	fnameCount := int(binary.LittleEndian.Uint32(name4Bin[0:4]))
	fnameStrData := fnameBin[4:] // après le premier int32

	type nameEntry struct {
		name string
	}
	names := make([]nameEntry, fnameCount)
	for i := 0; i < fnameCount; i++ {
		idxOff := 4 + i*8
		charOffset := int(binary.LittleEndian.Uint32(name4Bin[idxOff:]))
		charSize := int(binary.LittleEndian.Uint32(name4Bin[idxOff+4:]))
		byteStart := charOffset * 2
		byteEnd := byteStart + charSize*2
		if byteEnd > len(fnameStrData) {
			return fmt.Errorf("name4 entry %d out of bounds", i)
		}
		names[i].name = utf16LEToString(fnameStrData[byteStart:byteEnd])
	}

	// Créer le fichier PCK de sortie
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output PCK: %w", err)
	}
	defer out.Close()

	// Écrire un header vide pour le moment (on le remplira à la fin)
	hdr := &pckHeader{
		HdrLen:   92,
		Encrypt2: 1,
		Wtf:      wtfVal,
	}
	hdrBytes := makeHeader(hdr)
	out.Write(hdrBytes)
	pos := int64(92)

	// Écrire les sections binaires
	sectionPairs := make([]pairVal, len(sectionNames))
	for i, data := range sections {
		sectionPairs[i].Offset = int32(pos)
		sectionPairs[i].Size = int32(binary.LittleEndian.Uint32(data[0:4]))
		payload := data[4:]
		out.Write(payload)
		pos += int64(len(payload))
	}

	// Écrire la TOC (placeholder)
	tocOffset := pos
	filetoc := make([]byte, fnameCount*8)
	out.Write(filetoc)
	pos += int64(len(filetoc))

	// Écrire les données des fichiers
	dataOffset := pos
	tocEntries := make([]pairVal, fnameCount)

	for i, ne := range names {
		ssPath := filepath.Join(inputDir, ne.name+".ss")
		ssData, err := os.ReadFile(ssPath)
		if err != nil {
			return fmt.Errorf("cannot read %s.ss: %w", ne.name, err)
		}

		fmt.Printf("  [%d/%d] %s\n", i+1, fnameCount, ne.name+".ss")

		compData := Compress(ssData)
		// Chiffrement (même opération que déchiffrement — XOR est symétrique)
		decrypt(compData, 1, key1)

		tocEntries[i].Offset = int32(pos - dataOffset)
		tocEntries[i].Size = int32(len(compData))

		out.Write(compData)
		pos += int64(len(compData))
	}

	// Mettre à jour le header
	setSectionPairs(hdr, sectionPairs)
	hdr.FileToc.Offset = int32(tocOffset)
	hdr.FileToc.Size = int32(fnameCount)
	hdr.Data.Offset = int32(dataOffset)
	hdr.Data.Size = int32(fnameCount)

	// Réécrire le header
	out.Seek(0, 0)
	out.Write(makeHeader(hdr))

	// Réécrire la TOC
	out.Seek(tocOffset, 0)
	tocBuf := make([]byte, fnameCount*8)
	for i, te := range tocEntries {
		binary.LittleEndian.PutUint32(tocBuf[i*8:], uint32(te.Offset))
		binary.LittleEndian.PutUint32(tocBuf[i*8+4:], uint32(te.Size))
	}
	out.Write(tocBuf)

	fmt.Printf("PCK rebuilt → %s\n", outputPath)
	return nil
}

// ─── helpers ───────────────────────────────────────────────

func readHeader(buf []byte, hdr *pckHeader) error {
	if len(buf) < 92 {
		return fmt.Errorf("buffer too small for PCK header")
	}
	r := buf
	hdr.HdrLen = int32(binary.LittleEndian.Uint32(r[0:]))
	pairs := []*pairVal{
		&hdr.Table1, &hdr.GlobalVar, &hdr.GlobalVarStr,
		&hdr.Name1, &hdr.Name2, &hdr.Name3, &hdr.Name4,
		&hdr.FnameStr, &hdr.FileToc, &hdr.Data,
	}
	for i, p := range pairs {
		off := 4 + i*8
		p.Offset = int32(binary.LittleEndian.Uint32(r[off:]))
		p.Size = int32(binary.LittleEndian.Uint32(r[off+4:]))
	}
	hdr.Encrypt2 = int32(binary.LittleEndian.Uint32(r[84:]))
	hdr.Wtf = int32(binary.LittleEndian.Uint32(r[88:]))
	return nil
}

func makeHeader(hdr *pckHeader) []byte {
	b := make([]byte, 92)
	binary.LittleEndian.PutUint32(b[0:], uint32(hdr.HdrLen))
	pairs := []pairVal{
		hdr.Table1, hdr.GlobalVar, hdr.GlobalVarStr,
		hdr.Name1, hdr.Name2, hdr.Name3, hdr.Name4,
		hdr.FnameStr, hdr.FileToc, hdr.Data,
	}
	for i, p := range pairs {
		off := 4 + i*8
		binary.LittleEndian.PutUint32(b[off:], uint32(p.Offset))
		binary.LittleEndian.PutUint32(b[off+4:], uint32(p.Size))
	}
	binary.LittleEndian.PutUint32(b[84:], uint32(hdr.Encrypt2))
	binary.LittleEndian.PutUint32(b[88:], uint32(hdr.Wtf))
	return b
}

func getSectionPairs(hdr *pckHeader) []pairVal {
	return []pairVal{
		hdr.Table1, hdr.GlobalVar, hdr.GlobalVarStr,
		hdr.Name1, hdr.Name2, hdr.Name3, hdr.Name4,
		hdr.FnameStr, hdr.FileToc, // sentinelle pour calcul tailles
	}
}

func setSectionPairs(hdr *pckHeader, pairs []pairVal) {
	hdr.Table1 = pairs[0]
	hdr.GlobalVar = pairs[1]
	hdr.GlobalVarStr = pairs[2]
	hdr.Name1 = pairs[3]
	hdr.Name2 = pairs[4]
	hdr.Name3 = pairs[5]
	hdr.Name4 = pairs[6]
	hdr.FnameStr = pairs[7]
}

// dumpSection sauvegarde une section binaire dans outputDir/<name>.bin
// Format : [int32 size][données de la section]
func dumpSection(outputDir, name string, buf []byte, cur, next pairVal) error {
	start := int(cur.Offset)
	end := int(next.Offset)
	if end > len(buf) {
		end = len(buf)
	}
	if start > len(buf) || start >= end {
		return nil
	}
	data := buf[start:end]
	out := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint32(out[0:], uint32(cur.Size))
	copy(out[4:], data)
	return os.WriteFile(filepath.Join(outputDir, name+".bin"), out, 0644)
}

// GameNameFromPath tente de deviner le jeu depuis le chemin du fichier PCK
func GameNameFromPath(path string) string {
	dir := filepath.Dir(path)
	parts := strings.Split(filepath.ToSlash(dir), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if len(parts[i]) > 3 {
			return parts[i]
		}
	}
	return ""
}
