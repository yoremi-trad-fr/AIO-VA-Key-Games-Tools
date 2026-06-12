package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct - GUI backend, calls lucksystem.exe via subprocess
type App struct {
	ctx        context.Context
	lucksystem string // path to lucksystem.exe
	mu         sync.Mutex
	cancelFunc context.CancelFunc // cancels the running subprocess
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.findLuckSystem()
}

// ───────────────────────────────────────
// Find lucksystem executable
// ───────────────────────────────────────

// binDir returns the path to the bin/ directory relative to the GUI executable.
func (a *App) binDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "bin"
	}
	return filepath.Join(filepath.Dir(exePath), "bin")
}

// findTool searches for a tool binary in bin/ (with and without .exe).
func (a *App) findTool(name string) string {
	binDir := a.binDir()
	// 1. bin/<name>.exe
	candidate := filepath.Join(binDir, name+".exe")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	// 2. bin/<name> (Linux/Mac)
	candidate = filepath.Join(binDir, name)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	// 3. Same dir as GUI exe
	exePath, _ := os.Executable()
	if exePath != "" {
		candidate = filepath.Join(filepath.Dir(exePath), name+".exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		candidate = filepath.Join(filepath.Dir(exePath), name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// 4. PATH
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return ""
}

func (a *App) findLuckSystem() {
	a.lucksystem = a.findTool("lucksystem")
}

// GetLuckSystemPath returns the detected path (for UI display)
func (a *App) GetLuckSystemPath() string {
	return a.lucksystem
}

// SetLuckSystemPath allows the user to manually set the path
func (a *App) SetLuckSystemPath() string {
	file, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Locate lucksystem executable",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Executable", Pattern: "*.exe;lucksystem"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	if err != nil || file == "" {
		return a.lucksystem
	}
	a.lucksystem = file
	return a.lucksystem
}

// ───────────────────────────────────────
// Game Presets — scan data/ folder
// ───────────────────────────────────────

// GamePreset holds auto-detected game configuration from the data/ folder
type GamePreset struct {
	Name       string `json:"name"`       // Display name (e.g. "AIR", "LB_EN")
	OpcodeFile string `json:"opcodeFile"` // Absolute path to OPCODE .txt
	PluginFile string `json:"pluginFile"` // Absolute path to .py plugin (may be empty)
	GameFlag   string `json:"gameFlag"`   // Value for -g flag (game name)
}

// ScanGameData scans the data/ folder next to lucksystem and returns available game presets.
// Convention:
//   - data/GAME.txt          → opcode at root level (AIR, KANON, HARMONIA...)
//   - data/GAME/OPCODE.txt   → opcode in subdirectory (LB_EN, SP)
//   - data/GAME.py           → plugin file (always at root level)
//   - data/base/             → excluded (base modules, not game configs)
func (a *App) ScanGameData() []GamePreset {
	if a.lucksystem == "" {
		return nil
	}

	dataDir := filepath.Join(filepath.Dir(a.lucksystem), "data")
	info, err := os.Stat(dataDir)
	if err != nil || !info.IsDir() {
		return nil
	}

	presets := []GamePreset{}
	seen := map[string]bool{} // track game names to avoid duplicates

	// 1) Scan subdirectories (e.g. data/LB_EN/OPCODE.txt, data/SP/OPCODE.txt)
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		if strings.EqualFold(dirName, "base") {
			continue // skip base/ module directory
		}

		// Look for .txt files inside the subdirectory
		subDir := filepath.Join(dataDir, dirName)
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		for _, sub := range subEntries {
			if sub.IsDir() || !strings.HasSuffix(strings.ToLower(sub.Name()), ".txt") {
				continue
			}
			gameName := dirName
			if seen[gameName] {
				continue
			}
			seen[gameName] = true

			preset := GamePreset{
				Name:       gameName,
				OpcodeFile: filepath.Join(subDir, sub.Name()),
				GameFlag:   gameName,
			}
			// Check for plugin at data/GAME.py
			pluginPath := filepath.Join(dataDir, gameName+".py")
			if _, err := os.Stat(pluginPath); err == nil {
				preset.PluginFile = pluginPath
			}
			presets = append(presets, preset)
		}
	}

	// 2) Scan root-level .txt files (e.g. data/AIR.txt, data/KANON.txt)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".txt") {
			continue
		}

		gameName := strings.TrimSuffix(name, filepath.Ext(name))
		if seen[gameName] {
			continue // already found via subdirectory
		}
		seen[gameName] = true

		preset := GamePreset{
			Name:       gameName,
			OpcodeFile: filepath.Join(dataDir, name),
			GameFlag:   gameName,
		}
		// Check for plugin at data/GAME.py
		pluginPath := filepath.Join(dataDir, gameName+".py")
		if _, err := os.Stat(pluginPath); err == nil {
			preset.PluginFile = pluginPath
		}
		presets = append(presets, preset)
	}

	// Sort alphabetically by name
	sort.Slice(presets, func(i, j int) bool {
		return strings.ToLower(presets[i].Name) < strings.ToLower(presets[j].Name)
	})

	return presets
}

// ───────────────────────────────────────
// Console Logging
// ───────────────────────────────────────

func (a *App) log(msg string) {
	wailsRuntime.EventsEmit(a.ctx, "log", msg)
}

func (a *App) logError(msg string) {
	wailsRuntime.EventsEmit(a.ctx, "log", "[ERROR] "+msg)
}

func (a *App) logOK(msg string) {
	wailsRuntime.EventsEmit(a.ctx, "log", "[OK] "+msg)
}

// ───────────────────────────────────────
// Run lucksystem subprocess
// ───────────────────────────────────────

// runLuckSystem executes lucksystem with given arguments, streaming output to console
func (a *App) runLuckSystem(args ...string) error {
	if a.lucksystem == "" {
		a.logError("lucksystem.exe not found! Place it next to the GUI or use Settings to locate it.")
		return fmt.Errorf("lucksystem not found")
	}

	// Log the command being executed
	a.log(fmt.Sprintf("> %s %s", filepath.Base(a.lucksystem), strings.Join(args, " ")))

	// Create a cancellable context for this subprocess
	ctx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	a.cancelFunc = cancel
	a.mu.Unlock()
	defer func() {
		cancel()
		a.mu.Lock()
		a.cancelFunc = nil
		a.mu.Unlock()
	}()

	cmd := exec.CommandContext(ctx, a.lucksystem, args...)

	// Hide the CMD window on Windows (no console popup during batch operations)
	hideWindow(cmd)

	// Capture stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.logError(fmt.Sprintf("stdout pipe: %v", err))
		return err
	}

	// Capture stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.logError(fmt.Sprintf("stderr pipe: %v", err))
		return err
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		a.logError(fmt.Sprintf("Failed to start: %v", err))
		return err
	}

	// Stream stdout/stderr with batched logging for performance
	done := make(chan struct{}, 2)

	streamLines := func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			a.log(scanner.Text())
		}
		done <- struct{}{}
	}

	go streamLines(stdout)
	go streamLines(stderr)

	// Wait for both goroutines
	<-done
	<-done

	// Wait for process to finish
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			// Cancelled by user — not an error
			a.log("[STOPPED] Process cancelled by user.")
			return fmt.Errorf("cancelled")
		}
		a.logError(fmt.Sprintf("Process exited with error: %v", err))
		return err
	}

	return nil
}

// StopProcess cancels the currently running subprocess (called from frontend)
func (a *App) StopProcess() {
	a.mu.Lock()
	cancel := a.cancelFunc
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// ───────────────────────────────────────
// File Dialogs (generic)
// ───────────────────────────────────────

func (a *App) SelectPakFile() string {
	file, _ := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select .PAK file",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "PAK Files (*.PAK)", Pattern: "*.PAK;*.pak"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	return file
}

func (a *App) SelectFile(title string, pattern string, desc string) string {
	file, _ := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: title,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: desc, Pattern: pattern},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	return file
}

func (a *App) SelectDirectory(title string) string {
	dir, _ := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: title,
	})
	return dir
}

func (a *App) SelectSaveFile(title string, defaultName string, pattern string, desc string) string {
	file, _ := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultName,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: desc, Pattern: pattern},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	return file
}

// SelectOutputPath opens a save dialog that accepts any filename (no extension enforcement).
// Returns the full path as typed by the user. Uses OpenFile dialog in directory mode
// so Windows cannot reject the filename for missing extension.
func (a *App) SelectOutputDir(title string) string {
	dir, _ := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: title,
	})
	return dir
}

// stripScriptPakSuffix removes trailing "SCRIPT.PAK" or "SCRIPT.pak" from a directory path.
// game.go automatically appends /SCRIPT.PAK/ to the import/export path,
// so if the user already selected a SCRIPT.PAK folder, it would be doubled.
func stripScriptPakSuffix(dir string) string {
	base := strings.ToUpper(filepath.Base(dir))
	if base == "SCRIPT.PAK" {
		return filepath.Dir(dir)
	}
	return dir
}

// ═══════════════════════════════════════
// SCRIPT DECOMPILE
// ═══════════════════════════════════════
// lucksystem script decompile -s PAK -c charset -O opcode -p plugin -g game -o outputdir

func (a *App) ScriptDecompile(pakFile, opcodeFile, pluginFile, charsetStr, outputDir, gameName string) string {
	if pakFile == "" || outputDir == "" {
		a.logError("SCRIPT.PAK and output directory are required")
		return "ERROR"
	}
	if charsetStr == "" {
		charsetStr = "UTF-8"
	}

	// Strip trailing SCRIPT.PAK from output dir (game.go adds it automatically)
	outputDir = stripScriptPakSuffix(outputDir)

	a.log("════════════════════════════════════════")
	a.log("  SCRIPT DECOMPILE")
	a.log("════════════════════════════════════════")

	args := []string{"script", "decompile", "-s", pakFile, "-c", charsetStr, "-o", outputDir}
	if opcodeFile != "" {
		args = append(args, "-O", opcodeFile)
	}
	if pluginFile != "" {
		args = append(args, "-p", pluginFile)
	}
	if gameName != "" {
		args = append(args, "-g", gameName)
	}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK("Script decompile completed")
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// SCRIPT COMPILE (IMPORT)
// ═══════════════════════════════════════
// lucksystem script import -s PAK -c charset -O opcode -p plugin -g game -i importdir -o output.PAK

func (a *App) ScriptCompile(pakFile, opcodeFile, pluginFile, charsetStr, importDir, outputPak, gameName string) string {
	if pakFile == "" || importDir == "" || outputPak == "" {
		a.logError("SCRIPT.PAK, translated folder, and output PAK are required")
		return "ERROR"
	}
	if charsetStr == "" {
		charsetStr = "UTF-8"
	}

	// Strip trailing SCRIPT.PAK from import dir (game.go adds it automatically)
	importDir = stripScriptPakSuffix(importDir)

	a.log("════════════════════════════════════════")
	a.log("  SCRIPT COMPILE (IMPORT)")
	a.log("════════════════════════════════════════")

	args := []string{"script", "import", "-s", pakFile, "-c", charsetStr, "-i", importDir, "-o", outputPak}
	if opcodeFile != "" {
		args = append(args, "-O", opcodeFile)
	}
	if pluginFile != "" {
		args = append(args, "-p", pluginFile)
	}
	if gameName != "" {
		args = append(args, "-g", gameName)
	}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("Script compile completed -> %s", outputPak))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// PAK EXTRACT
// ═══════════════════════════════════════
// lucksystem pak extract -i PAK -o listfile --all outputdir -c charset

func (a *App) PakExtract(pakFile, outputDir string) string {
	if pakFile == "" || outputDir == "" {
		a.logError("PAK file and output directory are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  PAK EXTRACT")
	a.log("════════════════════════════════════════")

	os.MkdirAll(outputDir, os.ModePerm)

	// Nom du fichier liste = <NomDuPak>_list.txt  (ex: SYSCG.PAK → SYSCG_list.txt)
	pakBase := strings.TrimSuffix(filepath.Base(pakFile), filepath.Ext(pakFile))
	listFile := filepath.Join(outputDir, pakBase+"_list.txt")
	a.log(fmt.Sprintf("List file: %s", listFile))

	args := []string{"pak", "extract", "-i", pakFile, "-o", listFile, "--all", outputDir}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("PAK extracted to %s", outputDir))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// PAK REPLACE
// ═══════════════════════════════════════
// Mode dossier : lucksystem pak replace -s PAK -i inputdir -o output.PAK
// Mode liste   : lucksystem pak replace -s PAK -i listfile -l -o output.PAK

func (a *App) PakReplace(pakSource, inputDir, listFile, outputPak string) string {
	if pakSource == "" || outputPak == "" {
		a.logError("Original PAK and output PAK are required")
		return "ERROR"
	}

	// Exactly one of inputDir or listFile must be set
	useList := listFile != ""
	useDir := inputDir != ""
	if !useList && !useDir {
		a.logError("Provide either a folder or a list file as input")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  PAK REPLACE")
	a.log("════════════════════════════════════════")

	var args []string
	if useList {
		a.log(fmt.Sprintf("Mode: list file → %s", listFile))
		args = []string{"pak", "replace", "-s", pakSource, "-i", listFile, "-l", "-o", outputPak}
	} else {
		a.log(fmt.Sprintf("Mode: directory → %s", inputDir))
		args = []string{"pak", "replace", "-s", pakSource, "-i", inputDir, "-o", outputPak}
	}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("PAK written -> %s", outputPak))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// PAK FONT EXTRACT
// ═══════════════════════════════════════
// lucksystem pak extract -i PAK -o listfile --all outputdir -c charset

func (a *App) PakFontExtract(pakFile, charsetStr, outputDir string) string {
	if pakFile == "" || outputDir == "" {
		a.logError("PAK file and output directory are required")
		return "ERROR"
	}
	if charsetStr == "" {
		charsetStr = "UTF-8"
	}

	a.log("════════════════════════════════════════")
	a.log("  PAK (FONT) EXTRACT")
	a.log("════════════════════════════════════════")

	os.MkdirAll(outputDir, os.ModePerm)

	pakBase := strings.TrimSuffix(filepath.Base(pakFile), filepath.Ext(pakFile))
	listFile := filepath.Join(outputDir, pakBase+"_list.txt")
	a.log(fmt.Sprintf("List file: %s", listFile))

	args := []string{"pak", "extract", "-i", pakFile, "-o", listFile, "--all", outputDir, "-c", charsetStr}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("PAK (Font) extracted to %s", outputDir))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// PAK FONT REPLACE
// ═══════════════════════════════════════
// Mode liste   : lucksystem pak replace -s PAK -i listfile --list -o output.PAK -c charset
// Mode dossier : lucksystem pak replace -s PAK -i inputdir  -o output.PAK -c charset

func (a *App) PakFontReplace(pakSource, charsetStr, inputDir, listFile, outputPak string) string {
	if pakSource == "" || outputPak == "" {
		a.logError("Original PAK and output PAK are required")
		return "ERROR"
	}
	useList := listFile != ""
	useDir := inputDir != ""
	if !useList && !useDir {
		a.logError("Provide either a list file or a folder as input")
		return "ERROR"
	}
	if charsetStr == "" {
		charsetStr = "UTF-8"
	}

	a.log("════════════════════════════════════════")
	a.log("  PAK (FONT) REPLACE")
	a.log("════════════════════════════════════════")

	var args []string
	if useList {
		a.log(fmt.Sprintf("Mode: list file → %s", listFile))
		args = []string{"pak", "replace", "-s", pakSource, "-i", listFile, "-l", "-o", outputPak, "-c", charsetStr}
	} else {
		a.log(fmt.Sprintf("Mode: directory → %s", inputDir))
		args = []string{"pak", "replace", "-s", pakSource, "-i", inputDir, "-o", outputPak, "-c", charsetStr}
	}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("PAK (Font) written -> %s", outputPak))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// FONT EXTRACT
// ═══════════════════════════════════════
// lucksystem font extract -s czfile -S infofile -o output.png -O charset.txt

func (a *App) FontExtract(czFile, infoFile, outputPng, outputCharset string) string {
	if czFile == "" || infoFile == "" || outputPng == "" {
		a.logError("Font CZ file, info file, and output PNG are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  FONT EXTRACT")
	a.log("════════════════════════════════════════")

	args := []string{"font", "extract", "-s", czFile, "-S", infoFile, "-o", outputPng}
	if outputCharset != "" {
		args = append(args, "-O", outputCharset)
	}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("Font extracted -> %s", outputPng))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// FONT EDIT
// ═══════════════════════════════════════
// lucksystem font edit -s cz -S info -f ttf -o outcz -O outinfo [-r] [-a] [-i idx] [-c charset]

func (a *App) FontEdit(czFile, infoFile, ttfFile, outputCz, outputInfo, charsetFile string, redraw, appendMode bool, startIndex int) string {
	if czFile == "" || infoFile == "" || ttfFile == "" || outputCz == "" {
		a.logError("Font CZ, info, TTF, and output CZ are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  FONT EDIT")
	a.log("════════════════════════════════════════")

	args := []string{"font", "edit", "-s", czFile, "-S", infoFile, "-f", ttfFile, "-o", outputCz}
	if outputInfo != "" {
		args = append(args, "-O", outputInfo)
	}
	if charsetFile != "" {
		args = append(args, "-c", charsetFile)
	}
	if redraw {
		args = append(args, "-r")
	}
	if appendMode {
		args = append(args, "-a")
	} else if !redraw && startIndex > 0 {
		args = append(args, "-i", fmt.Sprintf("%d", startIndex))
	}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("Font edited -> %s", outputCz))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// IMAGE EXPORT (single)
// ═══════════════════════════════════════
// lucksystem image export -i czfile -o output.png

func (a *App) ImageExport(czFile, outputPng string) string {
	if czFile == "" || outputPng == "" {
		a.logError("CZ input and PNG output are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  IMAGE EXPORT")
	a.log("════════════════════════════════════════")

	args := []string{"image", "export", "-i", czFile, "-o", outputPng}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("Image exported -> %s", outputPng))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// IMAGE IMPORT (single)
// ═══════════════════════════════════════
// lucksystem image import -s source.cz -i input.png -o output.cz

func (a *App) ImageImport(sourceCz, inputPng, outputCz string, fill bool) string {
	if sourceCz == "" || inputPng == "" || outputCz == "" {
		a.logError("Source CZ, input PNG, and output CZ are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  IMAGE IMPORT")
	a.log("════════════════════════════════════════")

	args := []string{"image", "import", "-s", sourceCz, "-i", inputPng, "-o", outputCz}
	if fill {
		args = append(args, "-f")
	}

	err := a.runLuckSystem(args...)
	if err != nil {
		return "ERROR"
	}

	a.logOK(fmt.Sprintf("Image imported -> %s", outputCz))
	a.log("════════════════════════════════════════")
	return "OK"
}

// ═══════════════════════════════════════
// IMAGE BATCH EXPORT (directory)
// ═══════════════════════════════════════
// Iterates over all CZ files in inputDir, converts each to PNG in outputDir

func (a *App) ImageBatchExport(inputDir, outputDir string) string {
	if inputDir == "" || outputDir == "" {
		a.logError("Input and output directories are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  IMAGE BATCH EXPORT (CZ → PNG)")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Input:  %s", inputDir))
	a.log(fmt.Sprintf("Output: %s", outputDir))
	a.log("────────────────────────────────────────")

	os.MkdirAll(outputDir, os.ModePerm)

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		a.logError(fmt.Sprintf("Cannot read directory: %v", err))
		return "ERROR"
	}

	count := 0
	errors := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip files that already have a known extension (not CZ files)
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".png" || ext == ".txt" || ext == ".json" || ext == ".xml" {
			continue
		}

		inFile := filepath.Join(inputDir, name)
		outFile := filepath.Join(outputDir, name+".png")

		a.log(fmt.Sprintf("  [%d] %s ...", count+1, name))
		args := []string{"image", "export", "-i", inFile, "-o", outFile}
		if err := a.runLuckSystem(args...); err != nil {
			errors++
		} else {
			count++
		}
	}

	result := fmt.Sprintf("%d images exported, %d errors", count, errors)
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

// ═══════════════════════════════════════
// IMAGE BATCH IMPORT (directory)
// ═══════════════════════════════════════
// For each PNG in inputDir, finds matching CZ in sourceDir, imports, saves to outputDir

func (a *App) ImageBatchImport(sourceDir, inputDir, outputDir string, fill bool) string {
	if sourceDir == "" || inputDir == "" || outputDir == "" {
		a.logError("Source CZ dir, input PNG dir, and output dir are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  IMAGE BATCH IMPORT (PNG → CZ)")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Source CZ: %s", sourceDir))
	a.log(fmt.Sprintf("Input PNG: %s", inputDir))
	a.log(fmt.Sprintf("Output:    %s", outputDir))
	a.log("────────────────────────────────────────")

	os.MkdirAll(outputDir, os.ModePerm)

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		a.logError(fmt.Sprintf("Cannot read directory: %v", err))
		return "ERROR"
	}

	count := 0
	errors := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.ToLower(filepath.Ext(name)) != ".png" {
			continue
		}

		// Derive original CZ name: "filename.png" -> "filename"
		czName := strings.TrimSuffix(name, filepath.Ext(name))
		sourceCz := filepath.Join(sourceDir, czName)
		inputPng := filepath.Join(inputDir, name)
		outputCz := filepath.Join(outputDir, czName)

		// Check source CZ exists
		if _, err := os.Stat(sourceCz); os.IsNotExist(err) {
			a.log(fmt.Sprintf("  [SKIP] %s (no matching CZ: %s)", name, czName))
			continue
		}

		a.log(fmt.Sprintf("  [%d] %s ...", count+1, name))
		args := []string{"image", "import", "-s", sourceCz, "-i", inputPng, "-o", outputCz}
		if fill {
			args = append(args, "-f")
		}
		if err := a.runLuckSystem(args...); err != nil {
			errors++
		} else {
			count++
		}
	}

	result := fmt.Sprintf("%d images imported, %d errors", count, errors)
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

// ═══════════════════════════════════════
// FILE SELECTION HELPERS (Dialogue)
// ═══════════════════════════════════════

func (a *App) SelectScriptTxtFile() string {
	file, _ := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select decompiled script (.txt)",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	return file
}

func (a *App) SelectTsvFile() string {
	file, _ := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select TSV dialogue file",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "TSV/Text Files (*.txt;*.tsv)", Pattern: "*.txt;*.tsv"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	return file
}

func (a *App) SelectSaveTsvFile(defaultName string) string {
	file, _ := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Save TSV dialogue file",
		DefaultFilename: defaultName,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "TSV/Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	return file
}

func (a *App) SelectSaveScriptFile(defaultName string) string {
	file, _ := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Save patched script file",
		DefaultFilename: defaultName,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	})
	return file
}

// ═══════════════════════════════════════
// DIALOGUE EXTRACT / IMPORT
// ═══════════════════════════════════════
// Internal Go functions — NO lucksystem subprocess.
// Parses decompiled script .txt files from LuckSystem
// to extract/inject translatable entries as TSV.
//
// Supported line types:
//   MESSAGE(...)  — dialogue lines (all LuckEngine games)
//   LOG_BEGIN(...) — log/title entries (e.g. AIR, CLANNAD)
//   SELECT(...)   — in-game choice options (Kanon and others)
//
// All three types are treated identically for extract/import: each line
// produces one TSV row with N language columns (the Nth quoted string in
// the line). SELECT lines contain $d-separated choices inside their
// quoted strings; these are kept verbatim in the TSV cell.
//
// Lines may be prefixed by a "labelN: " marker (e.g. "label22: SELECT (...)")
// which is stripped before recognition.
//
// The user picks columns by number (Lang 1, Lang 2, ...).
// Column assignment varies by game — user must verify.

// DialogueFormatInfo is returned by DialogueDetectFormat
type DialogueFormatInfo struct {
	Format  string `json:"format"`
	MaxCols int    `json:"maxCols"`
}

// stripLabelPrefix removes an optional "labelN: " (or "labelXX: ") prefix
// from a trimmed line and returns the remainder. Returns the original line
// if no label prefix is present.
func stripLabelPrefix(trimmed string) string {
	// Match: ^label[0-9]+:\s*
	if !strings.HasPrefix(trimmed, "label") {
		return trimmed
	}
	rest := trimmed[len("label"):]
	// Consume digits
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(rest) || rest[i] != ':' {
		return trimmed
	}
	rest = rest[i+1:]
	// Consume optional whitespace
	rest = strings.TrimLeft(rest, " \t")
	return rest
}

// isDialogueLine returns true if the line (after optional label prefix
// stripping) starts with MESSAGE, LOG_BEGIN, or SELECT.
func isDialogueLine(trimmed string) bool {
	rest := stripLabelPrefix(trimmed)
	return strings.HasPrefix(rest, "MESSAGE") ||
		strings.HasPrefix(rest, "LOG_BEGIN") ||
		strings.HasPrefix(rest, "SELECT")
}

// lineTag returns "MESSAGE", "LOG_BEGIN", or "SELECT" for tagging in the TSV.
func lineTag(trimmed string) string {
	rest := stripLabelPrefix(trimmed)
	if strings.HasPrefix(rest, "LOG_BEGIN") {
		return "LOG_BEGIN"
	}
	if strings.HasPrefix(rest, "SELECT") {
		return "SELECT"
	}
	return "MESSAGE"
}

// isSelectLine returns true if the line (after label stripping) is a SELECT.
func isSelectLine(trimmed string) bool {
	return strings.HasPrefix(stripLabelPrefix(trimmed), "SELECT")
}

// DialogueDetectFormat reads a decompiled script and detects the format.
// Scans MESSAGE and LOG_BEGIN lines, counts max quoted strings.
func (a *App) DialogueDetectFormat(scriptFile string) DialogueFormatInfo {
	result := DialogueFormatInfo{Format: "Unknown", MaxCols: 0}

	if scriptFile == "" {
		return result
	}

	data, err := os.ReadFile(scriptFile)
	if err != nil {
		a.logError(fmt.Sprintf("Cannot read file: %v", err))
		return result
	}

	lines := strings.Split(string(data), "\n")

	maxQuotes := 0
	msgCount := 0
	logCount := 0
	selCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !isDialogueLine(trimmed) {
			continue
		}
		rest := stripLabelPrefix(trimmed)
		switch {
		case strings.HasPrefix(rest, "LOG_BEGIN"):
			logCount++
		case strings.HasPrefix(rest, "SELECT"):
			selCount++
		default:
			msgCount++
		}
		quotes := len(extractQuotedStrings(trimmed))
		if quotes > maxQuotes {
			maxQuotes = quotes
		}
		if msgCount+logCount+selCount >= 50 {
			break
		}
	}

	if msgCount+logCount+selCount == 0 {
		result.Format = "No MESSAGE / LOG_BEGIN / SELECT found"
		return result
	}

	// Cap at 4 columns
	if maxQuotes > 4 {
		maxQuotes = 4
	}
	result.MaxCols = maxQuotes

	parts := []string{}
	if msgCount > 0 {
		parts = append(parts, fmt.Sprintf("%d MESSAGE", msgCount))
	}
	if logCount > 0 {
		parts = append(parts, fmt.Sprintf("%d LOG_BEGIN", logCount))
	}
	if selCount > 0 {
		parts = append(parts, fmt.Sprintf("%d SELECT", selCount))
	}
	result.Format = fmt.Sprintf("%d columns detected (%s sampled)", maxQuotes, strings.Join(parts, " + "))

	return result
}

// extractQuotedStrings extracts all quoted strings from a line.
// Handles escaped quotes inside strings.
func extractQuotedStrings(line string) []string {
	var results []string
	inQuote := false
	var current strings.Builder
	runes := []rune(line)

	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if !inQuote {
			if ch == '"' {
				inQuote = true
				current.Reset()
			}
		} else {
			if ch == '\\' && i+1 < len(runes) && runes[i+1] == '"' {
				current.WriteRune('"')
				i++ // skip escaped quote
			} else if ch == '"' {
				results = append(results, current.String())
				inQuote = false
			} else {
				current.WriteRune(ch)
			}
		}
	}
	return results
}

// DialogueExtractFile extracts MESSAGE + LOG_BEGIN entries from a single script file to TSV.
func (a *App) DialogueExtractFile(inputFile, outputFile string, cols []int) string {
	if inputFile == "" || outputFile == "" {
		a.logError("Input script and output TSV file are required")
		return "ERROR"
	}
	if len(cols) == 0 {
		a.logError("At least one column must be selected")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  DIALOGUE EXTRACT (single file)")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Input:  %s", inputFile))
	a.log(fmt.Sprintf("Output: %s", outputFile))
	colNames := make([]string, len(cols))
	for i, c := range cols {
		colNames[i] = fmt.Sprintf("Lang %d", c)
	}
	a.log(fmt.Sprintf("Columns: %s", strings.Join(colNames, ", ")))

	count, err := a.extractDialoguesFromFile(inputFile, outputFile, cols)
	if err != nil {
		a.logError(fmt.Sprintf("Error: %v", err))
		return "ERROR"
	}

	result := fmt.Sprintf("%d entries extracted", count)
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

// DialogueExtractBatch extracts MESSAGE + LOG_BEGIN entries from all .txt scripts in a folder.
func (a *App) DialogueExtractBatch(inputDir, outputDir string, cols []int) string {
	if inputDir == "" || outputDir == "" {
		a.logError("Input folder and output folder are required")
		return "ERROR"
	}
	if len(cols) == 0 {
		a.logError("At least one column must be selected")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  DIALOGUE EXTRACT (batch)")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Input:  %s", inputDir))
	a.log(fmt.Sprintf("Output: %s", outputDir))
	colNames := make([]string, len(cols))
	for i, c := range cols {
		colNames[i] = fmt.Sprintf("Lang %d", c)
	}
	a.log(fmt.Sprintf("Columns: %s", strings.Join(colNames, ", ")))

	os.MkdirAll(outputDir, 0755)

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		a.logError(fmt.Sprintf("Cannot read directory: %v", err))
		return "ERROR"
	}

	totalEntries := 0
	fileCount := 0
	errors := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".txt") {
			continue
		}
		// Skip files that are already extracted TSV (*.ext.txt)
		if strings.HasSuffix(strings.ToLower(e.Name()), ".ext.txt") {
			continue
		}

		inPath := filepath.Join(inputDir, e.Name())
		outName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())) + ".ext.txt"
		outPath := filepath.Join(outputDir, outName)

		count, err := a.extractDialoguesFromFile(inPath, outPath, cols)
		if err != nil {
			a.log(fmt.Sprintf("  [SKIP] %s: %v", e.Name(), err))
			errors++
			continue
		}
		if count > 0 {
			a.log(fmt.Sprintf("  [%d] %s → %s (%d entries)", fileCount+1, e.Name(), outName, count))
			totalEntries += count
			fileCount++
		}
	}

	result := fmt.Sprintf("%d files processed, %d entries total, %d errors", fileCount, totalEntries, errors)
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

// extractDialoguesFromFile does the actual extraction work.
// cols contains 1-based column indices (e.g. [1, 2] for Lang 1 and Lang 2).
//
// All line types (MESSAGE, LOG_BEGIN, SELECT) are treated identically:
// 1 line → 1 TSV row (ID = seqID). The Nth quoted string maps to Lang N.
// For SELECT lines the quoted string(s) contain $d-separated choices that
// are kept verbatim in the TSV cell.
func (a *App) extractDialoguesFromFile(inputFile, outputFile string, cols []int) (int, error) {
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return 0, fmt.Errorf("cannot read %s: %v", inputFile, err)
	}

	lines := strings.Split(string(data), "\n")

	// Build TSV header: ID | TAG | Lang N | Lang M | ...
	var sb strings.Builder
	sb.WriteString("ID\tTAG")
	for _, col := range cols {
		sb.WriteString(fmt.Sprintf("\tLang %d", col))
	}
	sb.WriteString("\n")

	count := 0
	seqID := 0 // sequential ID for stable matching
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !isDialogueLine(trimmed) {
			continue
		}

		seqID++
		tag := lineTag(trimmed)
		quoted := extractQuotedStrings(trimmed)

		sb.WriteString(fmt.Sprintf("%d\t%s", seqID, tag))
		for _, col := range cols {
			sb.WriteString("\t")
			idx := col - 1 // convert 1-based to 0-based
			if idx >= 0 && idx < len(quoted) {
				text := strings.ReplaceAll(quoted[idx], "\t", "\\t")
				text = strings.ReplaceAll(text, "\n", "\\n")
				text = strings.ReplaceAll(text, "\r", "")
				sb.WriteString(text)
			}
		}
		sb.WriteString("\n")
		count++
	}

	if count == 0 {
		return 0, nil
	}

	if err := os.WriteFile(outputFile, []byte(sb.String()), 0644); err != nil {
		return 0, fmt.Errorf("cannot write %s: %v", outputFile, err)
	}

	return count, nil
}

// DialogueImportFile re-injects a TSV column back into a single script file.
// targetCol is 1-based (Lang 1, Lang 2, etc.).
func (a *App) DialogueImportFile(scriptFile, tsvFile string, targetCol int, outputFile string) string {
	if scriptFile == "" || tsvFile == "" || outputFile == "" {
		a.logError("Script file, TSV file, and output file are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  DIALOGUE IMPORT (single file)")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Script: %s", scriptFile))
	a.log(fmt.Sprintf("TSV:    %s", tsvFile))
	a.log(fmt.Sprintf("Target: Lang %d (quoted string #%d)", targetCol, targetCol))
	a.log(fmt.Sprintf("Output: %s", outputFile))

	count, err := a.importDialoguesToFile(scriptFile, tsvFile, targetCol, outputFile)
	if err != nil {
		a.logError(fmt.Sprintf("Error: %v", err))
		return "ERROR"
	}

	result := fmt.Sprintf("%d entries injected", count)
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

// DialogueImportBatch re-injects TSV columns into all matching scripts in a folder.
func (a *App) DialogueImportBatch(scriptsDir, tsvDir string, targetCol int, outputDir string) string {
	if scriptsDir == "" || tsvDir == "" || outputDir == "" {
		a.logError("Scripts folder, TSV folder, and output folder are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  DIALOGUE IMPORT (batch)")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Scripts: %s", scriptsDir))
	a.log(fmt.Sprintf("TSV:     %s", tsvDir))
	a.log(fmt.Sprintf("Target:  Lang %d", targetCol))
	a.log(fmt.Sprintf("Output:  %s", outputDir))

	os.MkdirAll(outputDir, 0755)

	entries, err := os.ReadDir(tsvDir)
	if err != nil {
		a.logError(fmt.Sprintf("Cannot read TSV directory: %v", err))
		return "ERROR"
	}

	totalEntries := 0
	fileCount := 0
	errors := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".ext.txt") {
			continue
		}

		// Derive script name: SEEN0001.ext.txt → SEEN0001.txt
		scriptName := strings.TrimSuffix(name, ".ext.txt") + ".txt"
		if strings.HasSuffix(strings.ToLower(name), ".EXT.txt") {
			scriptName = strings.TrimSuffix(name, ".EXT.txt") + ".txt"
		}

		scriptPath := filepath.Join(scriptsDir, scriptName)
		tsvPath := filepath.Join(tsvDir, name)
		outPath := filepath.Join(outputDir, scriptName)

		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			a.log(fmt.Sprintf("  [SKIP] %s (no matching script: %s)", name, scriptName))
			continue
		}

		count, err := a.importDialoguesToFile(scriptPath, tsvPath, targetCol, outPath)
		if err != nil {
			a.log(fmt.Sprintf("  [WARN] %s: %v", name, err))
			errors++
			continue
		}

		a.log(fmt.Sprintf("  [%d] %s + %s → %s (%d replaced)", fileCount+1, scriptName, name, scriptName, count))
		totalEntries += count
		fileCount++
	}

	result := fmt.Sprintf("%d files processed, %d entries injected, %d errors", fileCount, totalEntries, errors)
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

// importDialoguesToFile does the actual import work.
// It reads the TSV, builds a map of seqID→translated text, then replaces
// the corresponding quoted string in each MESSAGE/LOG_BEGIN/SELECT line.
// All three line types are treated identically (monobloc replacement).
// targetCol is 1-based: Lang 1 = replace quoted string #0, Lang 2 = #1, etc.
func (a *App) importDialoguesToFile(scriptFile, tsvFile string, targetCol int, outputFile string) (int, error) {
	// --- Read TSV ---
	tsvData, err := os.ReadFile(tsvFile)
	if err != nil {
		return 0, fmt.Errorf("cannot read TSV: %v", err)
	}

	tsvLines := strings.Split(string(tsvData), "\n")
	if len(tsvLines) < 2 {
		return 0, fmt.Errorf("TSV file is empty or has no data rows")
	}

	// Parse header to find the target column.
	// New format: ID | TAG | Lang 1 | Lang 2 | ...
	// We look for "Lang N" where N == targetCol, OR fall back to column index.
	header := strings.Split(strings.TrimSpace(tsvLines[0]), "\t")
	targetTsvCol := -1
	targetHeader := fmt.Sprintf("Lang %d", targetCol)
	for i, col := range header {
		if strings.EqualFold(strings.TrimSpace(col), targetHeader) {
			targetTsvCol = i
			break
		}
	}
	// Fallback: also accept old-style named headers (JAP=col1→0, ENG=col1→1, CN→2)
	if targetTsvCol < 0 {
		oldNames := map[int][]string{
			1: {"JAP"},
			2: {"ENG"},
			3: {"CN"},
		}
		if names, ok := oldNames[targetCol]; ok {
			for i, col := range header {
				for _, name := range names {
					if strings.EqualFold(strings.TrimSpace(col), name) {
						targetTsvCol = i
						break
					}
				}
				if targetTsvCol >= 0 {
					break
				}
			}
		}
	}
	if targetTsvCol < 0 {
		return 0, fmt.Errorf("column '%s' not found in TSV header: %v", targetHeader, header)
	}

	// Build map: sequential_ID → translated text
	translations := make(map[int]string)
	for _, tsvLine := range tsvLines[1:] {
		tsvLine = strings.TrimSpace(tsvLine)
		if tsvLine == "" {
			continue
		}
		tsvCols := strings.Split(tsvLine, "\t")
		if len(tsvCols) <= targetTsvCol {
			continue
		}
		// First column is the sequential ID
		idStr := strings.TrimSpace(tsvCols[0])
		id := 0
		fmt.Sscanf(idStr, "%d", &id)
		if id <= 0 {
			continue
		}
		text := tsvCols[targetTsvCol]
		text = strings.ReplaceAll(text, "\\t", "\t")
		text = strings.ReplaceAll(text, "\\n", "\n")
		if text != "" {
			translations[id] = text
		}
	}

	// --- Read script and replace ---
	scriptData, err := os.ReadFile(scriptFile)
	if err != nil {
		return 0, fmt.Errorf("cannot read script: %v", err)
	}

	replaceIdx := targetCol - 1 // 1-based to 0-based

	scriptLines := strings.Split(string(scriptData), "\n")
	count := 0
	seqID := 0

	for i, line := range scriptLines {
		trimmed := strings.TrimSpace(line)
		if !isDialogueLine(trimmed) {
			continue
		}

		seqID++
		newText, ok := translations[seqID]
		if !ok || newText == "" {
			continue
		}

		replaced := replaceNthQuotedString(line, replaceIdx, newText)
		if replaced != line {
			scriptLines[i] = replaced
			count++
		}
	}

	output := strings.Join(scriptLines, "\n")
	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		return 0, fmt.Errorf("cannot write output: %v", err)
	}

	return count, nil
}

// replaceNthQuotedString replaces the Nth (0-based) quoted string in a line.
func replaceNthQuotedString(line string, n int, newText string) string {
	// Escape special chars in newText for reinsertion
	escaped := strings.ReplaceAll(newText, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")

	runes := []rune(line)
	quoteCount := 0
	result := make([]rune, 0, len(runes)+len(escaped))
	inQuote := false
	skipUntilClose := false

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if !inQuote {
			if ch == '"' {
				if quoteCount == n {
					// This is the opening quote of the target string
					result = append(result, '"')
					result = append(result, []rune(escaped)...)
					// Skip until closing quote
					skipUntilClose = true
					inQuote = true
					continue
				}
				inQuote = true
				result = append(result, ch)
			} else {
				result = append(result, ch)
			}
		} else {
			if skipUntilClose {
				// Skip original content until we find the unescaped closing quote
				if ch == '\\' && i+1 < len(runes) {
					i++ // skip escaped char
					continue
				}
				if ch == '"' {
					result = append(result, '"')
					inQuote = false
					skipUntilClose = false
					quoteCount++
				}
				continue
			}
			// Normal pass-through of non-target quoted strings
			if ch == '\\' && i+1 < len(runes) && runes[i+1] == '"' {
				result = append(result, ch, runes[i+1])
				i++
				continue
			}
			if ch == '"' {
				result = append(result, ch)
				inQuote = false
				quoteCount++
			} else {
				result = append(result, ch)
			}
		}
	}

	return string(result)
}

// ─────────────────────────────────────────────────────────────
// Generic tool runner (reused for all bin/ tools)
// ─────────────────────────────────────────────────────────────

func (a *App) runTool(toolName string, args ...string) error {
	toolPath := a.findTool(toolName)
	if toolPath == "" {
		a.logError(fmt.Sprintf("%s not found! Place it in the bin/ folder.", toolName))
		return fmt.Errorf("%s not found", toolName)
	}

	a.log(fmt.Sprintf("> %s %s", toolName, strings.Join(args, " ")))

	ctx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	a.cancelFunc = cancel
	a.mu.Unlock()
	defer func() {
		cancel()
		a.mu.Lock()
		a.cancelFunc = nil
		a.mu.Unlock()
	}()

	cmd := exec.CommandContext(ctx, toolPath, args...)
	hideWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.logError(fmt.Sprintf("stdout pipe: %v", err))
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.logError(fmt.Sprintf("stderr pipe: %v", err))
		return err
	}

	if err := cmd.Start(); err != nil {
		a.logError(fmt.Sprintf("Failed to start %s: %v", toolName, err))
		return err
	}

	done := make(chan struct{}, 2)
	streamLines := func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			a.log(scanner.Text())
		}
		done <- struct{}{}
	}
	go streamLines(stdout)
	go streamLines(stderr)
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			a.log("[STOPPED] Process cancelled by user.")
			return fmt.Errorf("cancelled")
		}
		a.logError(fmt.Sprintf("Process exited with error: %v", err))
		return err
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// RLdev 2026 — Backend methods
// ─────────────────────────────────────────────────────────────

func (a *App) RldevDisassemble(seenFile, kfnFile, encoding, gameId string, debugInfo bool, outputDir string) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — Extract SEEN.txt")
	a.log("═══════════════════════════════════════")

	if err := required("SEEN.txt", seenFile); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}
	if encoding == "" {
		encoding = "UTF-8"
	}
	if kfnFile == "" {
		kfnFile = a.findKFN()
	}
	if kfnFile != "" {
		a.log("KFN: " + kfnFile)
	} else {
		a.log("Warning: reallive.kfn not found — opcodes will be raw")
	}

	args := []string{"-d", "-v", "1", "-e", encoding, "-o", outputDir}
	if kfnFile != "" {
		args = append(args, "-kfn", kfnFile)
	}
	if gameId != "" {
		args = append(args, "-G", gameId)
	}
	if debugInfo {
		args = append(args, "-g")
		a.log("Sources debug RealLive: yes (-g / #line)")
	}
	args = append(args, seenFile)

	if err := a.runTool("kprl16", args...); err != nil {
		return err.Error()
	}
	a.logOK("Extraction complete.")
	return ""
}

// findKFN searches for reallive.kfn in standard locations.
func (a *App) findKFN() string {
	binDir := a.binDir()
	candidates := []string{
		filepath.Join(binDir, "lib", "reallive.kfn"),
		filepath.Join(binDir, "reallive.kfn"),
	}
	exePath, _ := os.Executable()
	if exePath != "" {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "lib", "reallive.kfn"),
			filepath.Join(exeDir, "reallive.kfn"),
		)
		dir := exeDir
		for i := 0; i < 4 && dir != ""; i++ {
			candidates = append(candidates, filepath.Join(dir, "KFN", "reallive.kfn"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	rldev := os.Getenv("RLDEV")
	if rldev != "" {
		candidates = append(candidates,
			filepath.Join(rldev, "lib", "reallive.kfn"),
			filepath.Join(rldev, "reallive.kfn"),
		)
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "KFN", "reallive.kfn"))
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func (a *App) DefaultKFN() string {
	return a.findKFN()
}

func (a *App) DefaultBabelRoot() string {
	var candidates []string
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "BABEL"),
			filepath.Join(filepath.Dir(wd), "ResCODEX", "Rldev2026-go", "BABEL"),
			filepath.Join(filepath.Dir(filepath.Dir(wd)), "ResCODEX", "Rldev2026-go", "BABEL"),
		)
	}

	exePath, _ := os.Executable()
	if exePath != "" {
		dir := filepath.Dir(exePath)
		for i := 0; i < 5 && dir != ""; i++ {
			candidates = append(candidates,
				filepath.Join(dir, "BABEL"),
				filepath.Join(dir, "ResCODEX", "Rldev2026-go", "BABEL"),
			)
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	for _, candidate := range candidates {
		if isBabelRoot(candidate) {
			return candidate
		}
	}
	return ""
}

func isBabelRoot(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	if info, err := os.Stat(filepath.Join(path, "rtl", "rlBabel.dll")); err != nil || info.IsDir() {
		return false
	}
	if info, err := os.Stat(filepath.Join(path, "rtl", "rlBabelF.dll")); err != nil || info.IsDir() {
		return false
	}
	return true
}

func required(label string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", label)
	}
	return nil
}

func (a *App) failIf(err error) string {
	if err != nil {
		a.logError(err.Error())
		return err.Error()
	}
	return ""
}

var realLiveInterpreterCandidates = []string{
	"RealLive.exe",
	"RealLiveEn.exe",
	"Kinetic.exe",
	"kinetic.exe",
	"AVG2000.exe",
	"avg2000.exe",
	"SiglusEngine.exe",
	"siglusengine.exe",
	"SiglusEngine_Steam.exe",
	"siglusengine_steam.exe",
	"reallive.exe",
}

func findInterpreterInDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return ""
	}
	for _, name := range realLiveInterpreterCandidates {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

func (a *App) resolveInterpreter(gameexe, interpreter string) string {
	interpreter = strings.TrimSpace(interpreter)
	if interpreter != "" {
		return interpreter
	}

	gameexe = strings.TrimSpace(gameexe)
	if gameexe == "" {
		return ""
	}

	dir := gameexe
	if info, err := os.Stat(gameexe); err == nil {
		if !info.IsDir() {
			dir = filepath.Dir(gameexe)
		}
	} else if filepath.Ext(gameexe) != "" {
		dir = filepath.Dir(gameexe)
	}

	if found := findInterpreterInDir(dir); found != "" {
		a.logOK("Interpreteur detecte: " + found)
		return found
	}
	return ""
}

func (a *App) RldevExtract(seenFile, outputDir string) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — Extract SEEN.txt")
	a.log("═══════════════════════════════════════")
	if err := required("SEEN.txt", seenFile); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}
	if err := a.runTool("kprl16", "-x", "-o", outputDir, seenFile); err != nil {
		return err.Error()
	}
	a.logOK("Extraction complete.")
	return ""
}

func (a *App) RldevArchive(outputSeen, inputDir, templateSeen string) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — Create Archive")
	a.log("═══════════════════════════════════════")
	if err := required("output SEEN.txt", outputSeen); err != nil {
		return a.failIf(err)
	}
	if err := required("input folder", inputDir); err != nil {
		return a.failIf(err)
	}
	seen := map[string]bool{}
	var files []string
	for _, pattern := range []string{"*.TXT", "*.txt", "*.AVG", "*.avg"} {
		matches, _ := filepath.Glob(filepath.Join(inputDir, pattern))
		for _, file := range matches {
			key := strings.ToLower(file)
			if !seen[key] {
				seen[key] = true
				files = append(files, file)
			}
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		a.logError("No .TXT or .avg files found in " + inputDir)
		return "no .TXT or .avg files found"
	}
	args := []string{"-a"}
	if strings.TrimSpace(templateSeen) != "" {
		args = append(args, "-template", templateSeen)
		a.log("Template SEEN.txt: " + templateSeen)
	}
	args = append(args, outputSeen)
	args = append(args, files...)
	if err := a.runTool("kprl16", args...); err != nil {
		return err.Error()
	}
	a.logOK(fmt.Sprintf("Archive created with %d files.", len(files)))
	return ""
}

func (a *App) RldevList(seenFile string) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — List Archive Contents")
	a.log("═══════════════════════════════════════")
	if err := required("SEEN.txt", seenFile); err != nil {
		return a.failIf(err)
	}
	if err := a.runTool("kprl16", "-l", seenFile); err != nil {
		return err.Error()
	}
	return ""
}

func appendTransformArgs(args []string, transform string, forceTransform bool) []string {
	transform = strings.TrimSpace(transform)
	hasTransform := transform != "" && !strings.EqualFold(transform, "NONE")
	if hasTransform {
		args = append(args, "-x", transform)
	}
	if hasTransform && forceTransform {
		args = append(args, "--force-transform")
	}
	return args
}

func (a *App) RldevCompile(orgFile, kfnFile, gameexe, interpreter, targetVersion, encoding, transform string, forceTransform bool, outputDir string) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — Compile .org")
	a.log("═══════════════════════════════════════")

	if err := required("script .org/.ke/.avg", orgFile); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}

	if isAVG32SourceFile(orgFile) {
		if err := a.compileAVG32Source(orgFile, outputDir, transform, forceTransform); err != nil {
			return err.Error()
		}
		a.logOK("AVG32 compilation complete.")
		return ""
	}

	if encoding == "" {
		encoding = "UTF-8"
	}
	if kfnFile == "" {
		kfnFile = a.findKFN()
	}
	if err := required("KFN", kfnFile); err != nil {
		return a.failIf(err)
	}
	interpreter = a.resolveInterpreter(gameexe, interpreter)

	args := []string{"-v", "-e", encoding, "-d", outputDir}
	args = appendTransformArgs(args, transform, forceTransform)
	args = append(args, "-K", kfnFile)
	if gameexe != "" {
		args = append(args, "-i", gameexe)
	}
	if interpreter != "" {
		args = append(args, "-I", interpreter)
	}
	targetVersion = strings.TrimSpace(targetVersion)
	if targetVersion != "" {
		args = append(args, "--target-version", targetVersion)
		a.log("Version RealLive forced: " + targetVersion)
	}
	args = append(args, orgFile)
	if err := a.runTool("rlc2026", args...); err != nil {
		return err.Error()
	}
	a.logOK("Compilation complete.")
	return ""
}

// ═══════════════════════════════════════
// RLDEV COMPILE BATCH (directory)
// ═══════════════════════════════════════
// For each .org or .ke in inputDir, compiles to .TXT in outputDir.
// Mirrors the old shell scripts:
//
//   Clannad (Shift-JIS):
//     rlc -o SEENxxxx -d outdir -e cp932 -i gameexe.ini SEENxxxx.ke
//
//   Tomoyo / Western (UTF-8):
//     rlc -x Western -o SEENxxxx -d outdir -i gameexe.ini SEENxxxx.org
//     (encoding defaults to UTF-8 when -e is omitted)
//
// The output filename is derived from the input basename (without
// extension), exactly like the original .bat / .sh scripts.

func (a *App) RldevCompileBatch(inputDir, kfnFile, gameexe, interpreter, targetVersion, encoding, transform string, forceTransform bool, outputDir string) string {
	a.log("════════════════════════════════════════")
	a.log("  RLdev — COMPILE BATCH (.org/.ke/.avg → .TXT)")
	a.log("════════════════════════════════════════")
	if err := required("input folder", inputDir); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}
	a.log(fmt.Sprintf("Input:    %s", inputDir))
	a.log(fmt.Sprintf("Output:   %s", outputDir))
	if encoding == "" {
		encoding = "UTF-8"
	}
	a.log(fmt.Sprintf("Encoding: %s", encoding))
	if transform != "" {
		a.log(fmt.Sprintf("Transform: %s", transform))
	}
	a.log("────────────────────────────────────────")

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		a.logError(fmt.Sprintf("Cannot create output directory: %v", err))
		return "ERROR"
	}

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		a.logError(fmt.Sprintf("Cannot read directory: %v", err))
		return "ERROR"
	}

	// Collect .org, .ke and .avg files, sort for deterministic order.
	var sources []string
	hasKepago := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".org" || ext == ".ke" || ext == ".avg" {
			sources = append(sources, entry.Name())
			if ext == ".org" || ext == ".ke" {
				hasKepago = true
			}
		}
	}
	sort.Strings(sources)

	if len(sources) == 0 {
		a.logError("No .org, .ke or .avg files found in input directory")
		return "ERROR"
	}
	if hasKepago {
		if kfnFile == "" {
			kfnFile = a.findKFN()
		}
		if err := required("KFN", kfnFile); err != nil {
			return a.failIf(err)
		}
		interpreter = a.resolveInterpreter(gameexe, interpreter)
	}

	a.log(fmt.Sprintf("Found %d source file(s) to compile.", len(sources)))
	a.log("────────────────────────────────────────")

	count := 0
	errors := 0
	for i, name := range sources {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		inFile := filepath.Join(inputDir, name)

		a.log(fmt.Sprintf("  [%d/%d] %s ...", i+1, len(sources), name))

		if isAVG32SourceFile(inFile) {
			if err := a.compileAVG32Source(inFile, outputDir, transform, forceTransform); err != nil {
				errors++
				a.logError(fmt.Sprintf("    failed: %v", err))
			} else {
				count++
			}
			continue
		}

		args := []string{"-v", "-e", encoding, "-d", outputDir, "-o", base}
		args = appendTransformArgs(args, transform, forceTransform)
		args = append(args, "-K", kfnFile)
		if gameexe != "" {
			args = append(args, "-i", gameexe)
		}
		if interpreter != "" {
			args = append(args, "-I", interpreter)
		}
		targetVersion = strings.TrimSpace(targetVersion)
		if targetVersion != "" {
			args = append(args, "--target-version", targetVersion)
		}
		args = append(args, inFile)

		if err := a.runTool("rlc2026", args...); err != nil {
			errors++
			a.logError(fmt.Sprintf("    failed: %v", err))
		} else {
			count++
		}
	}

	result := fmt.Sprintf("%d file(s) compiled, %d error(s)", count, errors)
	if errors > 0 {
		a.logError(result)
	} else {
		a.logOK(result)
	}
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

func isAVG32SourceFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".avg")
}

func (a *App) compileAVG32Source(avgFile, outputDir, transform string, forceTransform bool) error {
	args := []string{"-c", "-t", "AVG32", "-v", "1", "-o", outputDir}
	args = appendKPRLTransformArgs(args, transform, forceTransform)
	args = append(args, avgFile)
	return a.runTool("kprl16", args...)
}

func appendKPRLTransformArgs(args []string, transform string, forceTransform bool) []string {
	transform = strings.TrimSpace(transform)
	hasTransform := transform != "" && !strings.EqualFold(transform, "NONE")
	if hasTransform {
		args = append(args, "-transform-output", transform)
	}
	if hasTransform && forceTransform {
		args = append(args, "-force-transform")
	}
	return args
}

func g00BatchFiles(inputDir, ext string) ([]string, error) {
	seen := map[string]bool{}
	var files []string
	for _, suffix := range []string{strings.ToLower(ext), strings.ToUpper(ext)} {
		matches, err := filepath.Glob(filepath.Join(inputDir, "*"+suffix))
		if err != nil {
			return nil, err
		}
		for _, file := range matches {
			if !seen[file] {
				seen[file] = true
				files = append(files, file)
			}
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no %s files found in %s", ext, inputDir)
	}
	return files, nil
}

func assetBatchFilesAny(inputDir string, exts ...string) ([]string, error) {
	seen := map[string]bool{}
	var files []string
	for _, ext := range exts {
		for _, suffix := range []string{strings.ToLower(ext), strings.ToUpper(ext)} {
			matches, err := filepath.Glob(filepath.Join(inputDir, "*"+suffix))
			if err != nil {
				return nil, err
			}
			for _, file := range matches {
				if !seen[file] {
					seen[file] = true
					files = append(files, file)
				}
			}
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no %s files found in %s", strings.Join(exts, "/"), inputDir)
	}
	return files, nil
}

func appendG00MetadataArg(args []string, xmlPath string) []string {
	xmlPath = strings.TrimSpace(xmlPath)
	if xmlPath != "" {
		args = append(args, "-m", xmlPath)
	}
	return args
}

func appendG00FormatArg(args []string, g00Format string) []string {
	g00Format = strings.TrimSpace(g00Format)
	if g00Format != "" && !strings.EqualFold(g00Format, "auto") {
		args = append(args, "-g", g00Format)
	}
	return args
}

func (a *App) RldevG00ToPng(g00Input, outputDir, xmlPath string, batch bool) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — G00 → PNG")
	a.log("═══════════════════════════════════════")
	label := "G00 file"
	if batch {
		label = "G00 folder"
	}
	if err := required(label, g00Input); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v", "-d", outputDir}
	args = appendG00MetadataArg(args, xmlPath)
	if batch {
		files, err := g00BatchFiles(g00Input, ".g00")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch G00: %d file(s)", len(files)))
		args = append(args, files...)
	} else {
		args = append(args, g00Input)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion complete.")
	return ""
}

func (a *App) RldevPngToG00(pngInput, outputDir, xmlPath, g00Format string, batch bool) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — PNG → G00")
	a.log("═══════════════════════════════════════")
	label := "PNG file"
	if batch {
		label = "PNG folder"
	}
	if err := required(label, pngInput); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v"}
	args = appendG00FormatArg(args, g00Format)
	args = appendG00MetadataArg(args, xmlPath)
	if batch {
		files, err := g00BatchFiles(pngInput, ".png")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch PNG: %d file(s)", len(files)))
		args = append(args, "-d", outputDir)
		args = append(args, files...)
	} else {
		base := strings.TrimSuffix(filepath.Base(pngInput), filepath.Ext(pngInput))
		outFile := filepath.Join(outputDir, base+".g00")
		args = append(args, "-o", outFile, "-i", pngInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion complete.")
	return ""
}

func (a *App) RldevGanToXml(ganFile, outputDir string) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — GAN → XML")
	a.log("═══════════════════════════════════════")
	if err := required("GAN file", ganFile); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}
	base := strings.TrimSuffix(filepath.Base(ganFile), filepath.Ext(ganFile))
	outFile := filepath.Join(outputDir, base+".ganxml")
	if err := a.runTool("rlxml", "-v", "-o", outFile, ganFile); err != nil {
		return err.Error()
	}
	a.logOK("Conversion complete: " + outFile)
	return ""
}

func (a *App) RldevXmlToGan(xmlFile, outputDir string) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — XML → GAN")
	a.log("═══════════════════════════════════════")
	if err := required("GANXML file", xmlFile); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}
	base := strings.TrimSuffix(filepath.Base(xmlFile), filepath.Ext(xmlFile))
	outFile := filepath.Join(outputDir, base+".gan")
	if err := a.runTool("rlxml", "-v", "-o", outFile, xmlFile); err != nil {
		return err.Error()
	}
	a.logOK("Conversion complete: " + outFile)
	return ""
}

func (a *App) RldevNwaToAudio(nwaInput, outputDir, audioFormat string, batch bool) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — NWA → audio")
	a.log("═══════════════════════════════════════")
	label := "NWA file"
	if batch {
		label = "NWA folder"
	}
	if err := required(label, nwaInput); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}

	audioFormat = strings.TrimSpace(strings.ToLower(audioFormat))
	if audioFormat == "" {
		audioFormat = "mp3"
	}

	args := []string{"-v", "-audio", audioFormat, "-d", outputDir}
	if batch {
		files, err := g00BatchFiles(nwaInput, ".nwa")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch NWA: %d file(s)", len(files)))
		args = append(args, files...)
	} else {
		args = append(args, nwaInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion complete.")
	return ""
}

func (a *App) RldevDatToJson(datInput, outputDir string, batch bool) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — CGM/TCC → JSON")
	a.log("═══════════════════════════════════════")
	label := "CGM/TCC file"
	if batch {
		label = "CGM/TCC folder"
	}
	if err := required(label, datInput); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v", "-d", outputDir}
	if batch {
		files, err := assetBatchFilesAny(datInput, ".cgm", ".tcc")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch DAT: %d file(s)", len(files)))
		args = append(args, files...)
	} else {
		args = append(args, datInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion complete.")
	return ""
}

func (a *App) RldevDatJsonToBinary(jsonInput, outputDir string, batch bool) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — JSON → CGM/TCC")
	a.log("═══════════════════════════════════════")
	label := "DAT JSON file"
	if batch {
		label = "DAT JSON folder"
	}
	if err := required(label, jsonInput); err != nil {
		return a.failIf(err)
	}
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v", "-d", outputDir}
	if batch {
		files, err := g00BatchFiles(jsonInput, ".json")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch JSON DAT: %d file(s)", len(files)))
		args = append(args, files...)
	} else {
		args = append(args, jsonInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion complete.")
	return ""
}

func (a *App) RldevBabelPrepareRuntime(babelRoot, gameDir, version, dllMode, nameEnc string, updateGameexe bool) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — Babel runtime setup")
	a.log("═══════════════════════════════════════")
	if err := required("BABEL folder", babelRoot); err != nil {
		return a.failIf(err)
	}
	if err := required("game folder", gameDir); err != nil {
		return a.failIf(err)
	}
	if !isBabelRoot(babelRoot) {
		return a.failIf(fmt.Errorf("invalid BABEL folder: %s", babelRoot))
	}
	if info, err := os.Stat(gameDir); err != nil || !info.IsDir() {
		return a.failIf(fmt.Errorf("invalid game folder: %s", gameDir))
	}

	version = strings.TrimSpace(version)
	dllName := resolveBabelDLLName(version, dllMode)
	srcDLL := filepath.Join(babelRoot, "rtl", dllName)
	dstDLL := filepath.Join(gameDir, dllName)
	if err := copyFile(srcDLL, dstDLL); err != nil {
		return a.failIf(err)
	}
	a.logOK("DLL copied: " + dstDLL)

	if version != "" {
		mapSrc := filepath.Join(babelRoot, "rtl", version+".map")
		if info, err := os.Stat(mapSrc); err == nil && !info.IsDir() {
			mapDst := filepath.Join(gameDir, version+".map")
			if err := copyFile(mapSrc, mapDst); err != nil {
				return a.failIf(err)
			}
			a.logOK("Map copied: " + mapDst)
		} else {
			a.log("Map not found for " + version + " (use rlbabel-genmap if this version is not bundled in the DLL).")
		}
	}

	if updateGameexe {
		gameexe := filepath.Join(gameDir, "GAMEEXE.INI")
		if err := updateBabelGameexe(gameexe, dllName, nameEnc); err != nil {
			return a.failIf(err)
		}
		a.logOK("GAMEEXE.INI updated: " + gameexe)
	} else {
		a.log("GAMEEXE.INI left untouched.")
	}

	if dllName == "rlBabelF.dll" {
		a.log("Note: rlBabelF is for older RealLive 1.2.x; load it at startup with LoadDLL(0, 'rlBabelF') or via rlcInit().")
	} else {
		a.log("Note: RealLive 1.2.5+ expects GAMEEXE to contain a #DLL.xxx = \"rlBabel\" line.")
	}
	a.logOK("Babel runtime setup complete.")
	return ""
}

func (a *App) RldevBabelWriteHeader(outputDir string, enableGlosses bool) string {
	a.log("═══════════════════════════════════════")
	a.log("  RLdev — Babel global.kh helper")
	a.log("═══════════════════════════════════════")
	if err := required("output folder", outputDir); err != nil {
		return a.failIf(err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return a.failIf(err)
	}

	var b strings.Builder
	b.WriteString("{- RLdev 2026 Babel helper -}\r\n")
	b.WriteString("#define __DynamicLineation__ = 1\r\n")
	if enableGlosses {
		b.WriteString("#define __EnableGlosses__\r\n")
	}
	b.WriteString("#load 'rlBabel'\r\n")
	path := filepath.Join(outputDir, "global.kh")
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return a.failIf(err)
	}
	a.logOK("Header created: " + path)
	a.log("Copy these lines at the beginning of the script to test, or into the project common header.")
	return ""
}

func resolveBabelDLLName(version, dllMode string) string {
	mode := strings.ToLower(strings.TrimSpace(dllMode))
	switch mode {
	case "old", "rlbabelf", "rlbabelf.dll":
		return "rlBabelF.dll"
	case "new", "rlbabel", "rlbabel.dll":
		return "rlBabel.dll"
	}
	if babelVersionBefore125(version) {
		return "rlBabelF.dll"
	}
	return "rlBabel.dll"
}

func babelVersionBefore125(version string) bool {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) < 3 {
		return false
	}
	nums := make([]int, 4)
	for i := 0; i < len(nums) && i < len(parts); i++ {
		fmt.Sscanf(parts[i], "%d", &nums[i])
	}
	if nums[0] != 1 {
		return false
	}
	if nums[1] < 2 {
		return true
	}
	if nums[1] > 2 {
		return false
	}
	return nums[2] < 5
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func updateBabelGameexe(path, dllName, nameEnc string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	backup := path + ".babel-" + time.Now().Format("20060102-150405") + ".bak"
	if err := os.WriteFile(backup, data, 0644); err != nil {
		return err
	}
	text := string(data)
	if dllName == "rlBabel.dll" && !regexp.MustCompile(`(?im)^#DLL\.\d{3}\s*=\s*"rlBabel"\s*$`).MatchString(text) {
		next := nextDLLSlot(text)
		text = appendGameexeLine(text, fmt.Sprintf("#DLL.%03d = \"rlBabel\"", next))
	}
	if encLine, ok := babelNameEncLine(nameEnc); ok {
		re := regexp.MustCompile(`(?im)^#NAME_ENC\s*=.*$`)
		if re.MatchString(text) {
			text = re.ReplaceAllString(text, encLine)
		} else {
			text = appendGameexeLine(text, encLine)
		}
	}
	return os.WriteFile(path, []byte(text), 0644)
}

func nextDLLSlot(text string) int {
	re := regexp.MustCompile(`(?im)^#DLL\.(\d{3})\s*=`)
	matches := re.FindAllStringSubmatch(text, -1)
	maxSlot := -1
	for _, m := range matches {
		var slot int
		if _, err := fmt.Sscanf(m[1], "%d", &slot); err == nil && slot > maxSlot {
			maxSlot = slot
		}
	}
	return maxSlot + 1
}

func appendGameexeLine(text, line string) string {
	if text != "" && !strings.HasSuffix(text, "\n") && !strings.HasSuffix(text, "\r") {
		text += "\r\n"
	}
	return text + line + "\r\n"
}

func babelNameEncLine(nameEnc string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(nameEnc)) {
	case "", "none":
		return "", false
	case "chinese", "1":
		return "#NAME_ENC = 1", true
	case "western", "2":
		return "#NAME_ENC = 2", true
	case "korean", "3":
		return "#NAME_ENC = 3", true
	default:
		return "", false
	}
}
