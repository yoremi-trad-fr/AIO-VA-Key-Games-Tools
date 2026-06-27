package siglus

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type mobilePCKHeader struct {
	Type       uint32
	FileCount  uint32
	DataOffset uint32
	SizeOffset uint32
}

// UnpackMobilePCK extracts the simple PCK archive used by mobile Siglus ports.
func UnpackMobilePCK(inputPath, outputDir string) error {
	buf, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("cannot read PCK: %w", err)
	}
	if len(buf) < 32 {
		return fmt.Errorf("file too small to be mobile PCK")
	}
	hdr := mobilePCKHeader{
		Type:       binary.LittleEndian.Uint32(buf[0:4]),
		FileCount:  binary.LittleEndian.Uint32(buf[4:8]),
		DataOffset: binary.LittleEndian.Uint32(buf[8:12]) + 32,
		SizeOffset: binary.LittleEndian.Uint32(buf[12:16]) + 32,
	}
	if hdr.Type != 1 {
		return fmt.Errorf("not a Siglus mobile PCK data file")
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output dir: %w", err)
	}

	count := int(hdr.FileCount)
	nameLenOffset := 32
	if nameLenOffset+count*4 > len(buf) {
		return fmt.Errorf("name length table out of bounds")
	}
	nameOffset := nameLenOffset + count*4
	names := make([]string, count)
	for i := 0; i < count; i++ {
		nameLen := int(binary.LittleEndian.Uint32(buf[nameLenOffset+i*4:]))
		if nameOffset+nameLen > len(buf) {
			return fmt.Errorf("filename %d out of bounds", i)
		}
		names[i] = utf16LEToString(buf[nameOffset : nameOffset+nameLen])
		nameOffset += nameLen
	}

	sizeOffset := int(hdr.SizeOffset)
	if sizeOffset+count*16 > len(buf) {
		return fmt.Errorf("file table out of bounds")
	}
	for i, name := range names {
		entryOff := sizeOffset + i*16
		fileOffset := int(binary.LittleEndian.Uint64(buf[entryOff:]))
		fileSize := int(binary.LittleEndian.Uint64(buf[entryOff+8:]))
		if fileOffset < 0 || fileSize < 0 || fileOffset+fileSize > len(buf) {
			return fmt.Errorf("file %s data out of bounds", name)
		}
		outPath, err := safeArchivePath(outputDir, name)
		if err != nil {
			return fmt.Errorf("unsafe filename in archive: %s", name)
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, buf[fileOffset:fileOffset+fileSize], 0644); err != nil {
			return fmt.Errorf("cannot write %s: %w", name, err)
		}
	}
	return nil
}

func safeArchivePath(outputDir, name string) (string, error) {
	cleanName := filepath.Clean(name)
	if cleanName == "." || cleanName == ".." || filepath.IsAbs(cleanName) || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe archive path")
	}
	outPath := filepath.Join(outputDir, cleanName)
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return "", err
	}
	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absOutput, absOutPath)
	if err != nil || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe archive path")
	}
	return outPath, nil
}

// PackMobilePCK packs direct children of inputDir into the mobile Siglus PCK format.
func PackMobilePCK(inputDir, outputPath string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("cannot read input dir: %w", err)
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no files to pack")
	}

	var nameTable bytes.Buffer
	nameLengths := make([]uint32, len(files))
	sizes := make([]int64, len(files))
	for i, name := range files {
		encodedName := stringToUTF16LE(name)
		nameLengths[i] = uint32(len(encodedName))
		nameTable.Write(encodedName)
		info, err := os.Stat(filepath.Join(inputDir, name))
		if err != nil {
			return err
		}
		sizes[i] = info.Size()
	}

	var archive bytes.Buffer
	archive.Write(make([]byte, 32))
	for _, n := range nameLengths {
		_ = binary.Write(&archive, binary.LittleEndian, n)
	}
	archive.Write(nameTable.Bytes())
	for archive.Len()%4 != 0 {
		archive.WriteByte(0)
	}

	sizeOffset := archive.Len()
	dataOffset := sizeOffset + len(files)*16
	offset := uint64(dataOffset)
	for _, size := range sizes {
		_ = binary.Write(&archive, binary.LittleEndian, offset)
		_ = binary.Write(&archive, binary.LittleEndian, uint64(size))
		offset += uint64(size)
	}

	header := archive.Bytes()
	binary.LittleEndian.PutUint32(header[0:4], 1)
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(files)))
	binary.LittleEndian.PutUint32(header[8:12], uint32(dataOffset-32))
	binary.LittleEndian.PutUint32(header[12:16], uint32(sizeOffset-32))

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output PCK: %w", err)
	}
	defer out.Close()
	if _, err := out.Write(header); err != nil {
		return err
	}
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(inputDir, name))
		if err != nil {
			return err
		}
		if _, err := out.Write(data); err != nil {
			return err
		}
	}
	return nil
}
