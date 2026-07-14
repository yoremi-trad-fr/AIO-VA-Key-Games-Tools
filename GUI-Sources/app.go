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
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"lucksystem/audio"
	"lucksystem/siglusluca"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct - GUI backend, calls lucksystem.exe via subprocess
type App struct {
	ctx        context.Context
	lucksystem string // path to lucksystem.exe
	mu         sync.Mutex
	cancelFunc context.CancelFunc // cancels the running subprocess
	logMu      sync.Mutex
	logFile    *os.File
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

// findTool searches for a bundled tool in common build/dev locations.
func (a *App) findTool(name string) string {
	var dirs []string
	addDir := func(dir string) {
		if dir == "" {
			return
		}
		dir = filepath.Clean(dir)
		for _, existing := range dirs {
			if strings.EqualFold(existing, dir) {
				return
			}
		}
		dirs = append(dirs, dir)
	}

	addDir(a.binDir())
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		addDir(exeDir)
		for dir, i := exeDir, 0; i < 5 && dir != ""; i++ {
			addDir(filepath.Join(dir, "bin"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if wd, err := os.Getwd(); err == nil {
		addDir(wd)
		for dir, i := wd, 0; i < 5 && dir != ""; i++ {
			addDir(filepath.Join(dir, "bin"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	names := []string{name + ".exe", name}
	for _, dir := range dirs {
		for _, fileName := range names {
			candidate := filepath.Join(dir, fileName)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
	}

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
	a.logMu.Lock()
	if a.logFile != nil {
		_, _ = fmt.Fprintln(a.logFile, msg)
	}
	a.logMu.Unlock()

	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", msg)
	}
}

func (a *App) logError(msg string) {
	a.log("[ERROR] " + msg)
}

func (a *App) logOK(msg string) {
	a.log("[OK] " + msg)
}

func (a *App) startLogFile(outputDir, prefix string) func() {
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return func() {}
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		a.logError(fmt.Sprintf("journal impossible: %v", err))
		return func() {}
	}

	name := fmt.Sprintf("%s-%s.log", prefix, time.Now().Format("20060102-150405"))
	path := filepath.Join(outputDir, name)
	file, err := os.Create(path)
	if err != nil {
		a.logError(fmt.Sprintf("journal impossible: %v", err))
		return func() {}
	}

	a.logMu.Lock()
	previous := a.logFile
	a.logFile = file
	a.logMu.Unlock()

	a.logOK("Log complet: " + path)

	return func() {
		a.logMu.Lock()
		if a.logFile == file {
			a.logFile = previous
		}
		a.logMu.Unlock()
		_ = file.Close()
	}
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
// SIGLUS -> LUCA SCRIPT BRIDGE
// ═══════════════════════════════════════
// Imports translated Siglus script text into decompiled Luca scripts.

func (a *App) SiglusLucaBridge(lucaDir, siglusDir, outputDir string, targetCol int) string {
	if lucaDir == "" || siglusDir == "" || outputDir == "" {
		a.logError("Luca scripts folder, Siglus Full folder, and output folder are required")
		return "ERROR"
	}
	if targetCol <= 0 {
		targetCol = 2
	}

	hdOutput := filepath.Join(outputDir, "hd_candidates.tsv")
	reviewOutput := filepath.Join(outputDir, "review.tsv")

	a.log("════════════════════════════════════════")
	a.log("  SIGLUS -> LUCA SCRIPT BRIDGE")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Luca scripts: %s", lucaDir))
	a.log(fmt.Sprintf("Siglus Full:  %s", siglusDir))
	a.log(fmt.Sprintf("Output:       %s", outputDir))
	a.log(fmt.Sprintf("Target col:   Lang %d", targetCol))

	summary, err := siglusluca.Run(siglusluca.Options{
		LucaDir:      lucaDir,
		SiglusDir:    siglusDir,
		OutputDir:    outputDir,
		HDOutput:     hdOutput,
		ReviewOutput: reviewOutput,
		TargetCol:    targetCol,
	})
	if err != nil {
		a.logError(err.Error())
		return "ERROR"
	}

	a.logOK("Siglus -> Luca bridge completed")
	a.log(fmt.Sprintf("  files processed: %d", summary.FilesProcessed))
	a.log(fmt.Sprintf("  files copied unchanged: %d", summary.FilesCopied))
	a.log(fmt.Sprintf("  imported lines: %d", summary.Imported))
	a.log(fmt.Sprintf("  HD candidate lines: %d", summary.HDCandidates))
	a.log(fmt.Sprintf("  review rows: %d", summary.ReviewRows))
	a.log(fmt.Sprintf("  low-confidence rows: %d", summary.LowConfidence))
	a.log(fmt.Sprintf("  output scripts: %s", outputDir))
	a.log(fmt.Sprintf("  HD candidates: %s", hdOutput))
	a.log(fmt.Sprintf("  review report: %s", reviewOutput))
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
// BGMOVIE / VIDEO EXTRACT
// ═══════════════════════════════════════
// Extracts BGMOVIE-style PAK files, then unwraps MVT movie files to WebM.

func (a *App) BGMOVIEExtract(pakFile, outputRoot string) string {
	if pakFile == "" || outputRoot == "" {
		a.logError("BGMOVIE.PAK and output directory are required")
		return "ERROR"
	}

	pakBase := strings.TrimSuffix(filepath.Base(pakFile), filepath.Ext(pakFile))
	extractDir := outputRoot
	if !strings.EqualFold(filepath.Base(outputRoot), pakBase) {
		extractDir = filepath.Join(outputRoot, pakBase)
	}
	webmDir := filepath.Join(extractDir, "webm")

	a.log("════════════════════════════════════════")
	a.log("  BGMOVIE / VIDEO EXTRACT")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("PAK:    %s", pakFile))
	a.log(fmt.Sprintf("Raw:    %s", extractDir))
	a.log(fmt.Sprintf("WebM:   %s", webmDir))
	a.log("────────────────────────────────────────")

	if err := os.MkdirAll(extractDir, os.ModePerm); err != nil {
		a.logError(fmt.Sprintf("Cannot create output directory: %v", err))
		return "ERROR"
	}

	listFile := filepath.Join(extractDir, pakBase+"_list.txt")
	args := []string{"pak", "extract", "-i", pakFile, "-o", listFile, "--all", extractDir}
	if err := a.runLuckSystem(args...); err != nil {
		return "ERROR"
	}

	if err := os.MkdirAll(webmDir, os.ModePerm); err != nil {
		a.logError(fmt.Sprintf("Cannot create WebM directory: %v", err))
		return "ERROR"
	}

	entries, err := os.ReadDir(extractDir)
	if err != nil {
		a.logError(fmt.Sprintf("Cannot read extracted directory: %v", err))
		return "ERROR"
	}

	movies := 0
	skipped := 0
	errors := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".txt" || ext == ".webm" {
			continue
		}

		inFile := filepath.Join(extractDir, name)
		if detectImageBatchAssetKind(inFile) != "mvt" {
			skipped++
			continue
		}

		outFile := filepath.Join(webmDir, name+".webm")
		a.log(fmt.Sprintf("  [%d] %s -> %s", movies+errors+1, name, filepath.Base(outFile)))
		if err := a.runLuckSystem("movie", "export", "-i", inFile, "-o", outFile); err != nil {
			errors++
		} else {
			movies++
		}
	}

	result := fmt.Sprintf("%d videos exported, %d skipped, %d errors", movies, skipped, errors)
	if movies == 0 && errors == 0 {
		a.logError("No MVT movie files found after PAK extraction")
		a.log("════════════════════════════════════════")
		return "ERROR"
	}
	if errors > 0 {
		a.logError(result)
		a.log("════════════════════════════════════════")
		return "ERROR"
	}

	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

// ═══════════════════════════════════════
// AUDIO PAK EXTRACT / CONVERT
// ═══════════════════════════════════════
// MUSIC/VOICE PAK entries are native Ogg Vorbis streams.

func (a *App) MusicPakExtract(pakFile, outputDir string, convertMP3 bool) string {
	return a.audioPakExtract("MUSIC", pakFile, outputDir, convertMP3)
}

func (a *App) VoicePakExtract(pakFile, outputDir string, convertMP3 bool) string {
	return a.audioPakExtract("VOICE", pakFile, outputDir, convertMP3)
}

func (a *App) audioPakExtract(kind, pakFile, outputDir string, convertMP3 bool) string {
	if pakFile == "" || outputDir == "" {
		a.logError(kind + " PAK and output directory are required")
		return "ERROR"
	}

	a.log("════════════════════════════════════════")
	a.log("  PAK AUDIO " + kind + " EXTRACT")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("PAK:    %s", pakFile))
	a.log(fmt.Sprintf("Output: %s", outputDir))
	if convertMP3 {
		a.log("MP3:    enabled")
	} else {
		a.log("MP3:    disabled")
	}
	a.log("────────────────────────────────────────")

	summary, err := audio.ExtractPak(audio.ExtractOptions{
		PakFile:    pakFile,
		OutputDir:  outputDir,
		Kind:       strings.ToLower(kind),
		ConvertMP3: convertMP3,
	}, func(line string) {
		a.log(line)
	})
	if err != nil {
		a.logError(err.Error())
		a.log("════════════════════════════════════════")
		return "ERROR"
	}

	a.log(fmt.Sprintf("List:   %s", summary.ListFile))
	a.log(fmt.Sprintf("Native: %s", summary.NativeDir))
	if summary.MP3Dir != "" {
		a.log(fmt.Sprintf("MP3:    %s", summary.MP3Dir))
	}
	result := fmt.Sprintf("%d Ogg extracted, %d MP3 converted, %d errors", summary.Files, summary.Converted, summary.Errors)
	if summary.Errors > 0 {
		a.logError(result)
		a.log("════════════════════════════════════════")
		return "ERROR"
	}
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK"
}

func (a *App) AudioConvert(inputPath, outputDir, direction string) string {
	if inputPath == "" || outputDir == "" {
		a.logError("Input audio file/folder and output directory are required")
		return "ERROR"
	}
	if direction == "" {
		direction = "mp3"
	}
	label := "Native Ogg -> MP3"
	if strings.EqualFold(direction, "native") || strings.EqualFold(direction, "ogg") {
		label = "MP3 -> Native Ogg"
	}

	a.log("════════════════════════════════════════")
	a.log("  AUDIO CONVERT")
	a.log("════════════════════════════════════════")
	a.log(fmt.Sprintf("Mode:   %s", label))
	a.log(fmt.Sprintf("Input:  %s", inputPath))
	a.log(fmt.Sprintf("Output: %s", outputDir))
	a.log("────────────────────────────────────────")

	summary, err := audio.ConvertPath(audio.ConvertOptions{
		InputPath: inputPath,
		OutputDir: outputDir,
		Direction: direction,
		Overwrite: true,
	}, func(line string) {
		a.log(line)
	})
	if err != nil {
		a.logError(err.Error())
		a.log("════════════════════════════════════════")
		return "ERROR"
	}

	result := fmt.Sprintf("%d converted, %d skipped, %d errors", summary.Converted, summary.Skipped, summary.Errors)
	if summary.Errors > 0 {
		a.logError(result)
		a.log("════════════════════════════════════════")
		return "ERROR"
	}
	a.logOK(result)
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
// Mode liste          : lucksystem pak replace -s PAK -i listfile --list -o output.PAK -c charset
// Mode dossier        : lucksystem pak replace -s PAK -i inputdir  -o output.PAK -c charset
// Mode fichier par nom: lucksystem pak replace -s PAK -i file --name internalName -o output.PAK -c charset
// Mode alias interne  : target-compatible font alias (keeps target info/CZ2 geometry)

func (a *App) PakFontReplace(pakSource, charsetStr, inputDir, listFile, singleFile, singleName, aliasFrom, aliasTo, outputPak string) string {
	if pakSource == "" || outputPak == "" {
		a.logError("Original PAK and output PAK are required")
		return "ERROR"
	}
	useList := listFile != ""
	useDir := inputDir != ""
	useSingle := singleFile != "" || singleName != ""
	useAlias := aliasFrom != "" || aliasTo != ""
	modeCount := 0
	if useList {
		modeCount++
	}
	if useDir {
		modeCount++
	}
	if useSingle {
		modeCount++
	}
	if useAlias {
		modeCount++
	}
	if modeCount != 1 {
		a.logError("Provide exactly one input mode: list, folder, single file by name, or internal alias")
		return "ERROR"
	}
	if useSingle && (singleFile == "" || singleName == "") {
		a.logError("Single-file mode requires both a file and an internal PAK name")
		return "ERROR"
	}
	if useAlias && (aliasFrom == "" || aliasTo == "") {
		a.logError("Alias mode requires both source and target internal names")
		return "ERROR"
	}
	if useAlias && strings.EqualFold(filepath.Clean(pakSource), filepath.Clean(outputPak)) {
		a.logError("Output PAK must be different from the original PAK")
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
	} else if useDir {
		a.log(fmt.Sprintf("Mode: directory → %s", inputDir))
		args = []string{"pak", "replace", "-s", pakSource, "-i", inputDir, "-o", outputPak, "-c", charsetStr}
	} else if useSingle {
		a.log(fmt.Sprintf("Mode: single file → %s as %s", singleFile, singleName))
		args = []string{"pak", "replace", "-s", pakSource, "-i", singleFile, "--name", singleName, "-o", outputPak, "-c", charsetStr}
	} else {
		a.log(fmt.Sprintf("Mode: target-compatible font alias → %s adapted to %s", aliasFrom, aliasTo))
		args = []string{"pak", "font-alias", "-s", pakSource, "--from-name", aliasFrom, "--to-name", aliasTo, "-o", outputPak, "-c", charsetStr}
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

func (a *App) FontEdit(czFile, infoFile, ttfFile, outputCz, outputInfo, charsetFile string, redraw, appendMode bool, startIndex int, arabicMetrics bool, metricSetYEnabled bool, metricSetY int, metricYOffset int, metricXOffset int, metricWOffset int, arabicConnectorBleed int) string {
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
	if arabicMetrics {
		args = append(args, "--arabic-metrics")
	}
	if metricSetYEnabled {
		args = append(args, "--metric-set-y", fmt.Sprintf("%d", metricSetY))
	}
	if metricYOffset != 0 {
		args = append(args, "--metric-y-offset", fmt.Sprintf("%d", metricYOffset))
	}
	if metricXOffset != 0 {
		args = append(args, "--metric-x-offset", fmt.Sprintf("%d", metricXOffset))
	}
	if metricWOffset != 0 {
		args = append(args, "--metric-w-offset", fmt.Sprintf("%d", metricWOffset))
	}
	if arabicConnectorBleed > 0 {
		args = append(args, "--arabic-connector-bleed", fmt.Sprintf("%d", arabicConnectorBleed))
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
// Iterates over supported image/movie files in inputDir and exports them.

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

	imageCount := 0
	movieCount := 0
	skipped := 0
	errors := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip files that already have a known extension (not CZ files)
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".png" || ext == ".webm" || ext == ".txt" || ext == ".json" || ext == ".xml" {
			continue
		}

		inFile := filepath.Join(inputDir, name)
		itemNumber := imageCount + movieCount + skipped + errors + 1

		switch detectImageBatchAssetKind(inFile) {
		case "cz":
			outFile := filepath.Join(outputDir, name+".png")
			a.log(fmt.Sprintf("  [%d] %s ...", itemNumber, name))
			args := []string{"image", "export", "-i", inFile, "-o", outFile}
			if err := a.runLuckSystem(args...); err != nil {
				errors++
			} else {
				imageCount++
			}
		case "mvt":
			outFile := filepath.Join(outputDir, name+".webm")
			a.log(fmt.Sprintf("  [%d] %s (MVT movie) ...", itemNumber, name))
			args := []string{"movie", "export", "-i", inFile, "-o", outFile}
			if err := a.runLuckSystem(args...); err != nil {
				errors++
			} else {
				movieCount++
			}
		default:
			a.log(fmt.Sprintf("  [SKIP] %s (not CZ/MVT)", name))
			skipped++
		}
	}

	result := fmt.Sprintf("%d images exported, %d movies exported, %d skipped, %d errors", imageCount, movieCount, skipped, errors)
	a.logOK(result)
	a.log("════════════════════════════════════════")
	return "OK: " + result
}

func detectImageBatchAssetKind(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer f.Close()

	var magic [4]byte
	n, _ := io.ReadFull(f, magic[:])
	if n >= 2 && magic[0] == 'C' && magic[1] == 'Z' {
		return "cz"
	}
	if n == 4 && magic[0] == 'M' && magic[1] == 'V' && magic[2] == 'T' && magic[3] == 0 {
		return "mvt"
	}
	return ""
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

		sourceCz, czName, candidates, ok := resolveImageBatchImportSource(sourceDir, name)
		inputPng := filepath.Join(inputDir, name)
		outputCz := filepath.Join(outputDir, czName)

		// Check source CZ exists
		if !ok {
			a.log(fmt.Sprintf("  [SKIP] %s (no matching CZ: %s)", name, strings.Join(candidates, " or ")))
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

func resolveImageBatchImportSource(sourceDir, pngName string) (sourceCz, outputName string, candidates []string, ok bool) {
	pngBase := strings.TrimSuffix(pngName, filepath.Ext(pngName))
	candidates = append(candidates, pngBase)

	switch strings.ToLower(filepath.Ext(pngBase)) {
	case ".cz0", ".cz1", ".cz2", ".cz3", ".cz4":
		withoutCzExt := strings.TrimSuffix(pngBase, filepath.Ext(pngBase))
		if withoutCzExt != "" && withoutCzExt != pngBase {
			candidates = append(candidates, withoutCzExt)
		}
	}

	for _, candidate := range candidates {
		path := filepath.Join(sourceDir, candidate)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, candidate, candidates, true
		}
	}
	return "", "", candidates, false
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
// RLdev 2026 — Backend methods
// ─────────────────────────────────────────────────────────────
func (a *App) executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("impossible de localiser l'executable: %w", err)
	}
	return filepath.Dir(exePath), nil
}

func (a *App) toolPath(toolName string) (string, error) {
	allowed := map[string]string{
		"kprl":       "kprl16",
		"kprl16":     "kprl16",
		"rlc":        "rlc2026",
		"rlc2026":    "rlc2026",
		"vaconv":     "vaconv",
		"rlxml":      "rlxml",
		"rlsave":     "rlsave",
		"siglustest": "siglustest",
	}

	binaryName, ok := allowed[toolName]
	if !ok {
		return "", fmt.Errorf("outil non pris en charge: %s", toolName)
	}

	path := a.findTool(binaryName)
	if path == "" {
		return "", fmt.Errorf("binaire manquant: %s (dossier bin ou PATH)", binaryName)
	}
	return path, nil
}
func (a *App) findKFN() string {
	var candidates []string

	binDir := a.binDir()
	candidates = append(candidates,
		filepath.Join(binDir, "lib", "reallive.kfn"),
		filepath.Join(binDir, "reallive.kfn"),
	)

	if exeDir, err := a.executableDir(); err == nil {
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

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "KFN", "reallive.kfn"))
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
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
	if exeDir, err := a.executableDir(); err == nil {
		dir := exeDir
		for i := 0; i < 5 && dir != ""; i++ {
			candidates = append(candidates, filepath.Join(dir, "BABEL"))
			candidates = append(candidates, filepath.Join(dir, "ResCODEX", "Rldev2026-go", "BABEL"))
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

func versionString(v [4]int) string {
	return fmt.Sprintf("%d.%d.%d.%d", v[0], v[1], v[2], v[3])
}

func peVersionFromExe(path string) ([4]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [4]int{}, err
	}
	if len(data) < 0x40 || data[0] != 'M' || data[1] != 'Z' {
		return [4]int{}, fmt.Errorf("ce n'est pas un executable PE")
	}
	peOff := int(uint32(data[0x3c]) | uint32(data[0x3d])<<8 | uint32(data[0x3e])<<16 | uint32(data[0x3f])<<24)
	if peOff+4 > len(data) || string(data[peOff:peOff+4]) != "PE\x00\x00" {
		return [4]int{}, fmt.Errorf("signature PE introuvable")
	}

	coffOff := peOff + 4
	if coffOff+20 > len(data) {
		return [4]int{}, fmt.Errorf("en-tete PE tronque")
	}
	nsec := int(uint16(data[coffOff+2]) | uint16(data[coffOff+3])<<8)
	optSize := int(uint16(data[coffOff+16]) | uint16(data[coffOff+17])<<8)
	secStart := coffOff + 20 + optSize

	rsrcOff := 0
	rsrcSize := 0
	for i := 0; i < nsec; i++ {
		s := secStart + i*40
		if s+40 > len(data) {
			break
		}
		name := strings.TrimRight(string(data[s:s+8]), "\x00")
		if name == ".rsrc" {
			rsrcSize = int(uint32(data[s+16]) | uint32(data[s+17])<<8 | uint32(data[s+18])<<16 | uint32(data[s+19])<<24)
			rsrcOff = int(uint32(data[s+20]) | uint32(data[s+21])<<8 | uint32(data[s+22])<<16 | uint32(data[s+23])<<24)
			break
		}
	}
	if rsrcOff == 0 {
		return [4]int{}, fmt.Errorf("section .rsrc introuvable")
	}
	end := rsrcOff + rsrcSize
	if end > len(data) {
		end = len(data)
	}

	idx := findVSFixedFileInfo(data, rsrcOff, end)
	if idx < 0 {
		idx = findVSFixedFileInfo(data, 0, len(data))
	}
	if idx < 0 {
		if v, err := peVersionFromStringFileInfo(data); err == nil {
			return v, nil
		}
		return [4]int{}, fmt.Errorf("version RealLive introuvable")
	}
	if idx+16 > len(data) {
		return [4]int{}, fmt.Errorf("version RealLive tronquee")
	}
	fvms := uint32(data[idx+8]) | uint32(data[idx+9])<<8 | uint32(data[idx+10])<<16 | uint32(data[idx+11])<<24
	fvls := uint32(data[idx+12]) | uint32(data[idx+13])<<8 | uint32(data[idx+14])<<16 | uint32(data[idx+15])<<24
	return [4]int{int(fvms >> 16), int(fvms & 0xffff), int(fvls >> 16), int(fvls & 0xffff)}, nil
}

func findVSFixedFileInfo(data []byte, start, end int) int {
	if start < 0 {
		start = 0
	}
	if end > len(data) {
		end = len(data)
	}
	if start >= end {
		return -1
	}
	sig := []byte{0xbd, 0x04, 0xef, 0xfe}
	for i := start; i+16 < end; i++ {
		if data[i] == sig[0] && data[i+1] == sig[1] && data[i+2] == sig[2] && data[i+3] == sig[3] {
			return i
		}
	}
	return -1
}

func peVersionFromStringFileInfo(data []byte) ([4]int, error) {
	key := utf16LEBytes("FileVersion")
	for i := 0; i+len(key) < len(data); i++ {
		if !bytesEqual(data[i:i+len(key)], key) {
			continue
		}
		searchEnd := i + len(key) + 256
		if searchEnd > len(data) {
			searchEnd = len(data)
		}
		for j := i + len(key); j+2 <= searchEnd; j += 2 {
			s := readUTF16LEString(data[j:searchEnd], 64)
			if v, ok := parseFileVersionString(s); ok {
				return v, nil
			}
		}
	}
	return [4]int{}, fmt.Errorf("FileVersion introuvable")
}

func utf16LEBytes(s string) []byte {
	words := utf16.Encode([]rune(s))
	out := make([]byte, 0, len(words)*2)
	for _, w := range words {
		out = append(out, byte(w), byte(w>>8))
	}
	return out
}

func readUTF16LEString(data []byte, maxRunes int) string {
	words := make([]uint16, 0, maxRunes)
	for i := 0; i+1 < len(data) && len(words) < maxRunes; i += 2 {
		w := uint16(data[i]) | uint16(data[i+1])<<8
		if w == 0 {
			break
		}
		words = append(words, w)
	}
	return string(utf16.Decode(words))
}

func parseFileVersionString(s string) ([4]int, bool) {
	if !strings.ContainsAny(s, ".,") {
		return [4]int{}, false
	}
	parts := make([]int, 0, 4)
	current := -1
	flush := func() bool {
		if current < 0 {
			return true
		}
		if current > 9999 {
			return false
		}
		parts = append(parts, current)
		current = -1
		return len(parts) <= 4
	}
	for _, r := range s {
		if r >= '0' && r <= '9' {
			if current < 0 {
				current = 0
			}
			current = current*10 + int(r-'0')
			continue
		}
		if !flush() {
			return [4]int{}, false
		}
	}
	if !flush() || len(parts) < 2 || len(parts) > 4 || parts[0] > 20 {
		return [4]int{}, false
	}
	var v [4]int
	for i, p := range parts {
		v[i] = p
	}
	return v, true
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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

func (a *App) DetectRealLiveVersion(gameexe, interpreter string) string {
	interpreter = a.resolveInterpreter(gameexe, interpreter)
	if interpreter == "" {
		return ""
	}
	version, err := peVersionFromExe(interpreter)
	if err != nil {
		a.log("Version RealLive non detectee: " + err.Error())
		return ""
	}
	text := versionString(version)
	a.logOK("Version RealLive detectee: " + text)
	return text
}

func (a *App) runTool(toolName string, args ...string) error {
	toolPath, err := a.toolPath(toolName)
	if err != nil {
		a.logError(err.Error())
		return err
	}

	a.log(fmt.Sprintf("> %s %s", filepath.Base(toolPath), strings.Join(args, " ")))

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
	cmd.Dir = filepath.Dir(toolPath)
	hideWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.logError(fmt.Sprintf("stdout: %v", err))
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.logError(fmt.Sprintf("stderr: %v", err))
		return err
	}

	if err := cmd.Start(); err != nil {
		a.logError(fmt.Sprintf("demarrage impossible: %v", err))
		return err
	}

	done := make(chan struct{}, 2)
	streamLines := func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			a.log(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			a.logError(fmt.Sprintf("lecture console: %v", err))
		}
		done <- struct{}{}
	}

	go streamLines(stdout)
	go streamLines(stderr)

	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			a.log("[STOPPED] Operation arretee par l'utilisateur.")
			return fmt.Errorf("operation arretee")
		}
		a.logError(fmt.Sprintf("processus termine en erreur: %v", err))
		return err
	}

	return nil
}

func required(label string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s est requis", label)
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

// ─────────────────────────────────────────────────────────────
// Siglus Tools 0.61 — Backend wrappers
// ─────────────────────────────────────────────────────────────

func siglusCompressionArgs(level int, fake bool) []string {
	if fake || level == 0 {
		return []string{"-f"}
	}
	if level < 2 {
		level = 2
	} else if level > 17 {
		level = 17
	}
	return []string{"-c", strconv.Itoa(level)}
}

func outputDirForFile(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.Dir(path)
}

func (a *App) SiglusSceneExtract(scenePck, gameName, outputDir string) string {
	if err := required("Scene.pck", scenePck); err != nil {
		return a.failIf(err)
	}
	if err := required("jeu", gameName); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "siglus-scene-extract")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "x", scenePck, gameName, outputDir))
}

func (a *App) SiglusSceneRebuild(inputDir, gameName, wtfVal, outputPck string, compressionLevel int, fakeCompression bool) string {
	if err := required("dossier .ss", inputDir); err != nil {
		return a.failIf(err)
	}
	if err := required("jeu", gameName); err != nil {
		return a.failIf(err)
	}
	if err := required("wtfval", wtfVal); err != nil {
		return a.failIf(err)
	}
	if err := required("Scene.pck de sortie", outputPck); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputPck), "siglus-scene-rebuild")
	defer closeLog()
	args := []string{"r", inputDir, gameName, wtfVal, outputPck}
	args = append(args, siglusCompressionArgs(compressionLevel, fakeCompression)...)
	return a.failIf(a.runTool("siglustest", args...))
}

func siglusSSDumpArgs(copyText, singleLine bool, filterMode, outputFormat string, singleXlsx bool) []string {
	var args []string
	if copyText {
		args = append(args, "-d")
	}
	if singleLine && strings.EqualFold(strings.TrimSpace(outputFormat), "txt") {
		args = append(args, "--single-line")
	}
	switch strings.ToLower(strings.TrimSpace(filterMode)) {
	case "all":
		args = append(args, "-a")
	case "dialogue", "dialogue-only", "filter-tags", "smart":
		args = append(args, "--dialogue-only")
	case "japanese", "japanese-only":
		args = append(args, "--japanese-only")
	case "full", "full-width", "fullwidth":
		args = append(args, "-w")
	}
	if strings.EqualFold(strings.TrimSpace(outputFormat), "xlsx") {
		args = append(args, "-x")
		if singleXlsx {
			args = append(args, "-s")
		}
	}
	return args
}

func (a *App) SiglusSSDump(ssFile, outputText string, copyText, singleLine bool, filterMode, outputFormat string) string {
	if err := required("fichier .ss", ssFile); err != nil {
		return a.failIf(err)
	}
	if err := required("texte de sortie", outputText); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputText), "siglus-ss-dump")
	defer closeLog()
	args := []string{"dump", ssFile, outputText}
	args = append(args, siglusSSDumpArgs(copyText, singleLine, filterMode, outputFormat, false)...)
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusSSDumpAll(ssDir, outputDir string, copyText, singleLine bool, filterMode, outputFormat string, singleXlsx bool) string {
	if err := required("dossier .ss", ssDir); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier texte", outputDir); err != nil {
		return a.failIf(err)
	}
	logDir := outputDir
	if strings.EqualFold(strings.TrimSpace(outputFormat), "xlsx") && singleXlsx {
		logDir = outputDirForFile(outputDir)
	}
	closeLog := a.startLogFile(logDir, "siglus-ss-dumpall")
	defer closeLog()
	args := []string{"dumpall", ssDir, outputDir}
	args = append(args, siglusSSDumpArgs(copyText, singleLine, filterMode, outputFormat, singleXlsx)...)
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusSSInject(ssFile, textFile, outputSS string) string {
	if err := required("fichier .ss original", ssFile); err != nil {
		return a.failIf(err)
	}
	if err := required("texte traduit", textFile); err != nil {
		return a.failIf(err)
	}
	if err := required(".ss de sortie", outputSS); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputSS), "siglus-ss-inject")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "inject", ssFile, textFile, outputSS))
}

func (a *App) SiglusSSInjectAll(ssDir, textDir, outputDir string) string {
	if err := required("dossier .ss original", ssDir); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier texte", textDir); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "siglus-ss-injectall")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "injectall", ssDir, textDir, outputDir))
}

func (a *App) SiglusGameexeExtract(gameexeDat, gameName, outputIni string) string {
	if err := required("Gameexe.dat", gameexeDat); err != nil {
		return a.failIf(err)
	}
	if err := required("jeu", gameName); err != nil {
		return a.failIf(err)
	}
	if err := required("INI de sortie", outputIni); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputIni), "siglus-gameexe-extract")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "gameexe-x", gameexeDat, gameName, outputIni))
}

func (a *App) SiglusGameexeRebuild(inputIni, gameName, outputDat string, doubleEncryption bool, compressionLevel int, fakeCompression bool) string {
	if err := required("Gameexe.ini", inputIni); err != nil {
		return a.failIf(err)
	}
	if err := required("jeu", gameName); err != nil {
		return a.failIf(err)
	}
	if err := required("Gameexe.dat de sortie", outputDat); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputDat), "siglus-gameexe-rebuild")
	defer closeLog()
	args := []string{"gameexe-r", inputIni, gameName, outputDat}
	if doubleEncryption {
		args = append(args, "-p")
	}
	args = append(args, siglusCompressionArgs(compressionLevel, fakeCompression)...)
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusDBSExtract(dbsFile, outputRaw, outputTxt, ansiEncoding string, dumpAll bool) string {
	if err := required("fichier .dbs", dbsFile); err != nil {
		return a.failIf(err)
	}
	if err := required(".dbs.out", outputRaw); err != nil {
		return a.failIf(err)
	}
	if err := required(".dbs.txt", outputTxt); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputTxt), "siglus-dbs-extract")
	defer closeLog()
	args := []string{"dbs-x", dbsFile, outputRaw, outputTxt}
	if dumpAll {
		args = append(args, "-a")
	}
	if strings.TrimSpace(ansiEncoding) != "" {
		args = append(args, "-e", ansiEncoding)
	}
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusDBSRebuild(rawFile, txtFile, outputDBS, ansiEncoding string, compressionLevel int, fakeCompression bool) string {
	if err := required(".dbs.out", rawFile); err != nil {
		return a.failIf(err)
	}
	if err := required(".dbs.txt", txtFile); err != nil {
		return a.failIf(err)
	}
	if err := required(".dbs de sortie", outputDBS); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputDBS), "siglus-dbs-rebuild")
	defer closeLog()
	args := []string{"dbs-r", rawFile, txtFile, outputDBS}
	args = append(args, siglusCompressionArgs(compressionLevel, fakeCompression)...)
	if strings.TrimSpace(ansiEncoding) != "" {
		args = append(args, "-e", ansiEncoding)
	}
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusDBSExportXLSX(dbsFile, outputXLSX, ansiEncoding string) string {
	if err := required("fichier .dbs", dbsFile); err != nil {
		return a.failIf(err)
	}
	if err := required("XLSX de sortie", outputXLSX); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputXLSX), "siglus-dbs-xlsx")
	defer closeLog()
	args := []string{"dbs-xlsx", dbsFile, outputXLSX}
	if strings.TrimSpace(ansiEncoding) != "" {
		args = append(args, "-e", ansiEncoding)
	}
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusDBSBuildFromXLSX(xlsxDir, outputDir, ansiEncoding string, unicodeOutput bool, compressionLevel int, fakeCompression bool) string {
	if err := required("dossier XLSX", xlsxDir); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier DBS de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "siglus-dbs-build")
	defer closeLog()
	args := []string{"dbs-build", xlsxDir, outputDir}
	if !unicodeOutput {
		args = append(args, "-e")
		if strings.TrimSpace(ansiEncoding) != "" {
			args = append(args, ansiEncoding)
		}
	}
	args = append(args, siglusCompressionArgs(compressionLevel, fakeCompression)...)
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusMobilePCKExtract(pckFile, outputDir string) string {
	if err := required("PCK mobile", pckFile); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "siglus-mobilepck-extract")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "mpck-x", pckFile, outputDir))
}

func (a *App) SiglusMobilePCKRebuild(inputDir, outputPCK string) string {
	if err := required("dossier PCK mobile", inputDir); err != nil {
		return a.failIf(err)
	}
	if err := required("PCK de sortie", outputPCK); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputPCK), "siglus-mobilepck-rebuild")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "mpck-r", inputDir, outputPCK))
}

func (a *App) SiglusOMVCut(inputOMV, outputOGV string) string {
	if err := required("OMV", inputOMV); err != nil {
		return a.failIf(err)
	}
	if err := required("OGV de sortie", outputOGV); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputOGV), "siglus-omv-cut")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "omv-cut", inputOMV, outputOGV))
}

func (a *App) SiglusOMVPack(inputOGV, outputOMV string) string {
	if err := required("OGV", inputOGV); err != nil {
		return a.failIf(err)
	}
	if err := required("OMV de sortie", outputOMV); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputOMV), "siglus-omv-pack")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "omv-pack", inputOGV, outputOMV))
}

func (a *App) SiglusOMV2AVI(inputOMV, outputFile string) string {
	if err := required("OMV", inputOMV); err != nil {
		return a.failIf(err)
	}
	if err := required("AVI / OGV de sortie", outputFile); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputFile), "siglus-omv2avi")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "omv2avi", inputOMV, outputFile))
}

func (a *App) SiglusOMVPNG(inputOMV, outputDir string) string {
	if err := required("OMV", inputOMV); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier PNG", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "siglus-omv-png")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "omv-png", inputOMV, outputDir))
}

func (a *App) SiglusPNGVideo(inputDir, outputFile string, alpha bool, fps string) string {
	if err := required("dossier PNG", inputDir); err != nil {
		return a.failIf(err)
	}
	if err := required("video de sortie", outputFile); err != nil {
		return a.failIf(err)
	}
	fps = strings.TrimSpace(fps)
	if fps == "" {
		fps = "30"
	}
	closeLog := a.startLogFile(outputDirForFile(outputFile), "siglus-png-video")
	defer closeLog()
	args := []string{"png-video", inputDir, outputFile, "--fps", fps}
	if alpha {
		args = append(args, "--alpha")
	}
	return a.failIf(a.runTool("siglustest", args...))
}

func (a *App) SiglusCombinePNG(inputDir, outputPNG string) string {
	if err := required("dossier PNG", inputDir); err != nil {
		return a.failIf(err)
	}
	if err := required("PNG de sortie", outputPNG); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputPNG), "siglus-combine-png")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "combine-png", inputDir, outputPNG))
}

func (a *App) SiglusScriptRepack(scriptFile, textFile, outputScript string) string {
	if err := required("script", scriptFile); err != nil {
		return a.failIf(err)
	}
	if err := required("texte UTF-16", textFile); err != nil {
		return a.failIf(err)
	}
	if err := required("script de sortie", outputScript); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDirForFile(outputScript), "siglus-script-repack")
	defer closeLog()
	return a.failIf(a.runTool("siglustest", "script-repack", scriptFile, textFile, outputScript))
}

func (a *App) RldevList(seenFile string) string {
	a.log("========================================")
	a.log("  RLdev - Liste SEEN.txt")
	a.log("========================================")

	if err := required("SEEN.txt", seenFile); err != nil {
		return a.failIf(err)
	}
	if err := a.runTool("kprl", "-l", seenFile); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RldevDisassemble(seenFile, kfnFile, encoding, gameID string, debugInfo bool, outputDir string) string {
	a.log("========================================")
	a.log("  RLdev - Desassemblage SEEN.txt")
	a.log("========================================")

	if err := required("SEEN.txt", seenFile); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "kprl-disasm")
	defer closeLog()

	if encoding == "" {
		encoding = "UTF-8"
	}
	if kfnFile == "" {
		kfnFile = a.findKFN()
	}
	if err := required("KFN", kfnFile); err != nil {
		return a.failIf(err)
	}

	args := []string{"-d", "-v", "1", "-e", encoding, "-o", outputDir}
	a.log("KFN: " + kfnFile)
	args = append(args, "-kfn", kfnFile)
	if gameID != "" {
		args = append(args, "-G", gameID)
	}
	if debugInfo {
		args = append(args, "-g")
		a.log("Sources debug RealLive: oui (-g / #line)")
	}
	args = append(args, seenFile)

	if err := a.runTool("kprl", args...); err != nil {
		return err.Error()
	}
	a.logOK("Desassemblage termine.")
	return ""
}

func (a *App) RldevExtract(seenFile, outputDir string) string {
	a.log("========================================")
	a.log("  RLdev - Extraction brute SEEN.txt")
	a.log("========================================")

	if err := required("SEEN.txt", seenFile); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	if err := a.runTool("kprl", "-x", "-v", "1", "-o", outputDir, seenFile); err != nil {
		return err.Error()
	}
	a.logOK("Extraction terminee.")
	return ""
}

func (a *App) RldevArchive(outputSeen, inputDir, templateSeen string) string {
	a.log("========================================")
	a.log("  RLdev - Reconstruction SEEN.txt")
	a.log("========================================")

	if err := required("SEEN.txt de sortie", outputSeen); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier d'entree", inputDir); err != nil {
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
		return a.failIf(fmt.Errorf("aucun fichier .TXT ou .avg trouve dans %s", inputDir))
	}

	args := []string{"-a"}
	if strings.TrimSpace(templateSeen) != "" {
		args = append(args, "-template", templateSeen)
		a.log("Template SEEN.txt: " + templateSeen)
	}
	args = append(args, outputSeen)
	args = append(args, files...)
	if err := a.runTool("kprl", args...); err != nil {
		return err.Error()
	}
	a.logOK(fmt.Sprintf("Archive reconstruite avec %d fichier(s).", len(files)))
	return ""
}

func appendTransformArgs(args []string, outputTransform string, forceTransform bool) []string {
	outputTransform = strings.TrimSpace(outputTransform)
	hasTransform := outputTransform != "" && !strings.EqualFold(outputTransform, "NONE")
	if hasTransform {
		args = append(args, "-x", outputTransform)
	}
	if hasTransform && forceTransform {
		args = append(args, "--force-transform")
	}
	return args
}

func (a *App) RldevCompile(orgFile, kfnFile, gameexe, interpreter, targetVersion, encoding, outputTransform string, forceTransform bool, outputDir string) string {
	a.log("========================================")
	a.log("  RLdev - Compilation script")
	a.log("========================================")

	if err := required("script .org/.ke/.avg", orgFile); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "rlc-compile")
	defer closeLog()

	if isAVG32SourceFile(orgFile) {
		if err := a.compileAVG32Source(orgFile, outputDir, outputTransform, forceTransform); err != nil {
			return err.Error()
		}
		a.logOK("Compilation AVG32 terminee.")
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
	args = appendTransformArgs(args, outputTransform, forceTransform)
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
		a.log("Version RealLive forcee: " + targetVersion)
	}
	args = append(args, orgFile)

	if err := a.runTool("rlc", args...); err != nil {
		return err.Error()
	}
	a.logOK("Compilation terminee.")
	return ""
}

func (a *App) RldevCompileBatch(inputDir, kfnFile, gameexe, interpreter, targetVersion, encoding, outputTransform string, forceTransform bool, outputDir string) string {
	a.log("========================================")
	a.log("  RLdev - Compilation batch scripts")
	a.log("========================================")

	if err := required("dossier d'entree", inputDir); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "rlc-batch")
	defer closeLog()

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return a.failIf(fmt.Errorf("lecture du dossier impossible: %w", err))
	}

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
		return a.failIf(fmt.Errorf("aucun fichier .org, .ke ou .avg trouve dans %s", inputDir))
	}

	if hasKepago {
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
	}

	okCount := 0
	errCount := 0
	for i, name := range sources {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		inputFile := filepath.Join(inputDir, name)
		a.log(fmt.Sprintf("[%d/%d] %s", i+1, len(sources), name))

		if isAVG32SourceFile(inputFile) {
			if err := a.compileAVG32Source(inputFile, outputDir, outputTransform, forceTransform); err != nil {
				errCount++
				a.logError(fmt.Sprintf("%s: %v", name, err))
				continue
			}
			okCount++
			continue
		}

		args := []string{"-v", "-e", encoding, "-d", outputDir, "-o", base}
		args = appendTransformArgs(args, outputTransform, forceTransform)
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
		args = append(args, inputFile)

		if err := a.runTool("rlc", args...); err != nil {
			errCount++
			a.logError(fmt.Sprintf("%s: %v", name, err))
			continue
		}
		okCount++
	}

	result := fmt.Sprintf("%d fichier(s) compile(s), %d erreur(s)", okCount, errCount)
	if errCount > 0 {
		a.logError(result)
		return result
	}
	a.logOK(result)
	return ""
}

func isAVG32SourceFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".avg")
}

func (a *App) compileAVG32Source(avgFile, outputDir, outputTransform string, forceTransform bool) error {
	args := []string{"-c", "-t", "AVG32", "-v", "1", "-o", outputDir}
	args = appendKPRLTransformArgs(args, outputTransform, forceTransform)
	args = append(args, avgFile)
	return a.runTool("kprl", args...)
}

func appendKPRLTransformArgs(args []string, outputTransform string, forceTransform bool) []string {
	outputTransform = strings.TrimSpace(outputTransform)
	hasTransform := outputTransform != "" && !strings.EqualFold(outputTransform, "NONE")
	if hasTransform {
		args = append(args, "-transform-output", outputTransform)
	}
	if hasTransform && forceTransform {
		args = append(args, "-force-transform")
	}
	return args
}

func orgTextBatchFiles(inputDir string) ([]string, error) {
	return assetBatchFilesAny(inputDir, ".org", ".ke")
}

func (a *App) RldevOrgTextExport(orgInput, outputDir, encoding string, batch bool) string {
	a.log("========================================")
	a.log("  RLdev - Export texte ORG/KE")
	a.log("========================================")

	label := "fichier .org/.ke"
	if batch {
		label = "dossier .org/.ke"
	}
	if err := required(label, orgInput); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "orgtext-export")
	defer closeLog()
	if encoding == "" {
		encoding = "UTF-8"
	}

	if batch {
		files, err := orgTextBatchFiles(orgInput)
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch ORG/KE: %d fichier(s)", len(files)))
		for i, file := range files {
			a.log(fmt.Sprintf("[%d/%d] %s", i+1, len(files), filepath.Base(file)))
			if err := a.runTool("rlc", "--text-export", "-e", encoding, "-d", outputDir, file); err != nil {
				return err.Error()
			}
		}
		a.logOK("Export texte termine.")
		return ""
	}

	if err := a.runTool("rlc", "--text-export", "-e", encoding, "-d", outputDir, orgInput); err != nil {
		return err.Error()
	}
	a.logOK("Export texte termine.")
	return ""
}

func (a *App) RldevOrgTextImport(orgInput, utfInput, outputDir, encoding string, batch bool) string {
	a.log("========================================")
	a.log("  RLdev - Import texte ORG/KE")
	a.log("========================================")

	orgLabel := "fichier .org/.ke"
	utfLabel := "fichier .utf"
	if batch {
		orgLabel = "dossier .org/.ke"
		utfLabel = "dossier .utf"
	}
	if err := required(orgLabel, orgInput); err != nil {
		return a.failIf(err)
	}
	if err := required(utfLabel, utfInput); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}
	closeLog := a.startLogFile(outputDir, "orgtext-import")
	defer closeLog()
	if encoding == "" {
		encoding = "UTF-8"
	}

	if batch {
		files, err := orgTextBatchFiles(orgInput)
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch ORG/KE: %d fichier(s)", len(files)))
		for i, file := range files {
			base := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			utfFile := filepath.Join(utfInput, base+".utf")
			if info, err := os.Stat(utfFile); err != nil || info.IsDir() {
				a.log(fmt.Sprintf("[SKIP] %s: .utf absent", filepath.Base(file)))
				continue
			}
			a.log(fmt.Sprintf("[%d/%d] %s", i+1, len(files), filepath.Base(file)))
			if err := a.runTool("rlc", "--text-import", "--text-file", utfFile, "-e", encoding, "-d", outputDir, file); err != nil {
				return err.Error()
			}
		}
		a.logOK("Import texte termine.")
		return ""
	}

	if err := a.runTool("rlc", "--text-import", "--text-file", utfInput, "-e", encoding, "-d", outputDir, orgInput); err != nil {
		return err.Error()
	}
	a.logOK("Import texte termine.")
	return ""
}

func assetBatchFiles(inputDir, ext string) ([]string, error) {
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
		return nil, fmt.Errorf("aucun fichier %s trouve dans %s", ext, inputDir)
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
		return nil, fmt.Errorf("aucun fichier %s trouve dans %s", strings.Join(exts, "/"), inputDir)
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
	return a.runG00ToPng("RLdev - G00 vers PNG", g00Input, outputDir, xmlPath, batch)
}

func (a *App) SiglusG00ToPng(g00Input, outputDir, xmlPath string, batch bool) string {
	return a.runG00ToPng("Siglus / vaconv - G00 vers PNG", g00Input, outputDir, xmlPath, batch)
}

func (a *App) runG00ToPng(title, g00Input, outputDir, xmlPath string, batch bool) string {
	a.log("========================================")
	a.log("  " + title)
	a.log("========================================")

	label := "fichier G00"
	if batch {
		label = "dossier G00"
	}
	if err := required(label, g00Input); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v", "-d", outputDir}
	args = appendG00MetadataArg(args, xmlPath)
	if batch {
		files, err := assetBatchFiles(g00Input, ".g00")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch G00: %d fichier(s)", len(files)))
		args = append(args, "-i", "g00", g00Input)
	} else {
		args = append(args, g00Input)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion terminee.")
	return ""
}

func (a *App) RldevPngToG00(pngInput, outputDir, xmlPath, g00Format string, batch bool) string {
	return a.runPngToG00("RLdev - PNG vers G00", pngInput, outputDir, xmlPath, g00Format, batch)
}

func (a *App) SiglusPngToG00(pngInput, outputDir, xmlPath, g00Format string, batch bool) string {
	return a.runPngToG00("Siglus / vaconv - PNG vers G00", pngInput, outputDir, xmlPath, g00Format, batch)
}

func (a *App) runPngToG00(title, pngInput, outputDir, xmlPath, g00Format string, batch bool) string {
	a.log("========================================")
	a.log("  " + title)
	a.log("========================================")

	label := "fichier PNG"
	if batch {
		label = "dossier PNG"
	}
	if err := required(label, pngInput); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v"}
	args = appendG00FormatArg(args, g00Format)
	args = appendG00MetadataArg(args, xmlPath)
	if batch {
		files, err := assetBatchFiles(pngInput, ".png")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch PNG: %d fichier(s)", len(files)))
		args = append(args, "-i", "png", "-d", outputDir, pngInput)
	} else {
		base := strings.TrimSuffix(filepath.Base(pngInput), filepath.Ext(pngInput))
		outputFile := filepath.Join(outputDir, base+".g00")
		args = append(args, "-o", outputFile, "-i", pngInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion terminee.")
	return ""
}

func (a *App) RldevGanToXml(ganFile, outputDir string) string {
	a.log("========================================")
	a.log("  RLdev - GAN vers XML")
	a.log("========================================")

	if err := required("fichier GAN", ganFile); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	base := strings.TrimSuffix(filepath.Base(ganFile), filepath.Ext(ganFile))
	outputFile := filepath.Join(outputDir, base+".ganxml")
	if err := a.runTool("rlxml", "-v", "-o", outputFile, ganFile); err != nil {
		return err.Error()
	}
	a.logOK("Conversion terminee: " + outputFile)
	return ""
}

func (a *App) RldevXmlToGan(xmlFile, outputDir string) string {
	a.log("========================================")
	a.log("  RLdev - XML vers GAN")
	a.log("========================================")

	if err := required("fichier GANXML", xmlFile); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	base := strings.TrimSuffix(filepath.Base(xmlFile), filepath.Ext(xmlFile))
	outputFile := filepath.Join(outputDir, base+".gan")
	if err := a.runTool("rlxml", "-v", "-o", outputFile, xmlFile); err != nil {
		return err.Error()
	}
	a.logOK("Conversion terminee: " + outputFile)
	return ""
}

func (a *App) RldevNwaToAudio(nwaInput, outputDir, audioFormat string, batch bool) string {
	a.log("========================================")
	a.log("  RLdev - NWA vers audio")
	a.log("========================================")

	label := "fichier NWA"
	if batch {
		label = "dossier NWA"
	}
	if err := required(label, nwaInput); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	audioFormat = strings.TrimSpace(strings.ToLower(audioFormat))
	if audioFormat == "" {
		audioFormat = "mp3"
	}

	args := []string{"-v", "-audio", audioFormat, "-d", outputDir}
	if batch {
		files, err := assetBatchFiles(nwaInput, ".nwa")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch NWA: %d fichier(s)", len(files)))
		args = append(args, "-i", "nwa", nwaInput)
	} else {
		args = append(args, nwaInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion terminee.")
	return ""
}

func (a *App) RldevDatToJson(datInput, outputDir string, batch bool) string {
	a.log("========================================")
	a.log("  RLdev - CGM/TCC vers JSON")
	a.log("========================================")

	label := "fichier CGM/TCC"
	if batch {
		label = "dossier CGM/TCC"
	}
	if err := required(label, datInput); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v", "-d", outputDir}
	if batch {
		files, err := assetBatchFilesAny(datInput, ".cgm", ".tcc")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch DAT: %d fichier(s)", len(files)))
		args = append(args, "-i", "dat", datInput)
	} else {
		args = append(args, datInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion terminee.")
	return ""
}

func (a *App) RldevDatJsonToBinary(jsonInput, outputDir string, batch bool) string {
	a.log("========================================")
	a.log("  RLdev - JSON vers CGM/TCC")
	a.log("========================================")

	label := "fichier JSON DAT"
	if batch {
		label = "dossier JSON DAT"
	}
	if err := required(label, jsonInput); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier de sortie", outputDir); err != nil {
		return a.failIf(err)
	}

	args := []string{"-v", "-d", outputDir}
	if batch {
		files, err := assetBatchFiles(jsonInput, ".json")
		if err != nil {
			return a.failIf(err)
		}
		a.log(fmt.Sprintf("Batch JSON DAT: %d fichier(s)", len(files)))
		args = append(args, "-i", "json", jsonInput)
	} else {
		args = append(args, jsonInput)
	}
	if err := a.runTool("vaconv", args...); err != nil {
		return err.Error()
	}
	a.logOK("Conversion terminee.")
	return ""
}

func (a *App) RldevSaveInfo(saveFile string) string {
	a.log("========================================")
	a.log("  RLdev - Infos sauvegarde RealLive")
	a.log("========================================")

	if err := required("fichier .sav", saveFile); err != nil {
		return a.failIf(err)
	}
	if err := a.runTool("rlsave", "info", saveFile); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RldevSaveMap(savePath string, jsonOutput bool) string {
	a.log("========================================")
	a.log("  RLdev - Cartographie sauvegardes RealLive")
	a.log("========================================")

	if err := required("fichier ou dossier .sav", savePath); err != nil {
		return a.failIf(err)
	}
	args := []string{"map"}
	if jsonOutput {
		args = append(args, "-json")
	}
	args = append(args, savePath)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RldevSaveDoctor(savePath string, jsonOutput bool) string {
	a.log("========================================")
	a.log("  RLdev - Diagnostic sauvegardes RealLive")
	a.log("========================================")

	if err := required("fichier ou dossier .sav", savePath); err != nil {
		return a.failIf(err)
	}
	args := []string{"doctor"}
	if jsonOutput {
		args = append(args, "-json")
	}
	args = append(args, savePath)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RldevSaveDiff(beforeSave, afterSave string, jsonOutput bool) string {
	a.log("========================================")
	a.log("  RLdev - Comparaison sauvegardes RealLive")
	a.log("========================================")

	if err := required("premiere sauvegarde .sav", beforeSave); err != nil {
		return a.failIf(err)
	}
	if err := required("deuxieme sauvegarde .sav", afterSave); err != nil {
		return a.failIf(err)
	}
	args := []string{"diff"}
	if jsonOutput {
		args = append(args, "-json")
	}
	args = append(args, beforeSave, afterSave)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RldevSaveGet(saveFile, refs string) string {
	a.log("========================================")
	a.log("  RLdev - Lecture sauvegarde RealLive")
	a.log("========================================")

	if err := required("fichier .sav", saveFile); err != nil {
		return a.failIf(err)
	}
	fields, err := saveArgFields("variables", refs)
	if err != nil {
		return a.failIf(err)
	}
	args := append([]string{"get", saveFile}, fields...)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RldevSaveSet(saveFile, assignments string, backup bool) string {
	a.log("========================================")
	a.log("  RLdev - Edition sauvegarde RealLive")
	a.log("========================================")

	if err := required("fichier .sav", saveFile); err != nil {
		return a.failIf(err)
	}
	fields, err := saveArgFields("assignations", assignments)
	if err != nil {
		return a.failIf(err)
	}
	args := []string{"set"}
	if !backup {
		args = append(args, "-no-backup")
	}
	args = append(args, saveFile)
	args = append(args, fields...)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	a.logOK("Sauvegarde mise a jour.")
	return ""
}

func (a *App) RldevSaveDump(saveFile string, includeAll, jsonOutput bool) string {
	a.log("========================================")
	a.log("  RLdev - Dump sauvegarde RealLive")
	a.log("========================================")

	if err := required("fichier .sav", saveFile); err != nil {
		return a.failIf(err)
	}
	args := []string{"dump"}
	if includeAll {
		args = append(args, "-all")
	}
	if jsonOutput {
		args = append(args, "-json")
	}
	args = append(args, saveFile)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) RldevSaveExport(saveFile, outputText string, lossless bool) string {
	a.log("========================================")
	a.log("  RLdev - Export texte sauvegarde RealLive")
	a.log("========================================")

	if err := required("fichier .sav", saveFile); err != nil {
		return a.failIf(err)
	}
	if err := required("fichier texte de sortie", outputText); err != nil {
		return a.failIf(err)
	}
	outputText = ensurePathExtension(outputText, ".txt")
	args := []string{"export"}
	if lossless {
		args = append(args, "-lossless")
	}
	args = append(args, saveFile, outputText)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	a.logOK("Export termine.")
	return ""
}

func ensurePathExtension(path, ext string) string {
	path = strings.TrimSpace(path)
	current := filepath.Ext(path)
	if current == "" || current == "." {
		return path + ext
	}
	return path
}

func (a *App) RldevSaveBuild(inputText, outputSave string, backup bool) string {
	a.log("========================================")
	a.log("  RLdev - Rebuild sauvegarde RealLive")
	a.log("========================================")

	if err := required("fichier texte", inputText); err != nil {
		return a.failIf(err)
	}
	if err := required("fichier .sav de sortie", outputSave); err != nil {
		return a.failIf(err)
	}
	args := []string{"build"}
	if !backup {
		args = append(args, "-no-backup")
	}
	args = append(args, inputText, outputSave)
	if err := a.runTool("rlsave", args...); err != nil {
		return err.Error()
	}
	a.logOK("Sauvegarde reconstruite.")
	return ""
}

func saveArgFields(label, value string) ([]string, error) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return nil, fmt.Errorf("%s requis", label)
	}
	return fields, nil
}

func (a *App) RldevBabelPrepareRuntime(babelRoot, gameDir, version, dllMode, nameEnc string, updateGameexe bool) string {
	a.log("========================================")
	a.log("  RLdev - Preparation runtime Babel")
	a.log("========================================")

	if err := required("dossier BABEL", babelRoot); err != nil {
		return a.failIf(err)
	}
	if err := required("dossier du jeu", gameDir); err != nil {
		return a.failIf(err)
	}
	if !isBabelRoot(babelRoot) {
		return a.failIf(fmt.Errorf("dossier BABEL invalide: %s", babelRoot))
	}
	if info, err := os.Stat(gameDir); err != nil || !info.IsDir() {
		return a.failIf(fmt.Errorf("dossier du jeu invalide: %s", gameDir))
	}

	version = strings.TrimSpace(version)
	dllName := resolveBabelDLLName(version, dllMode)
	srcDLL := filepath.Join(babelRoot, "rtl", dllName)
	dstDLL := filepath.Join(gameDir, dllName)
	if err := copyFile(srcDLL, dstDLL); err != nil {
		return a.failIf(err)
	}
	a.logOK("DLL copiee: " + dstDLL)

	if version != "" {
		mapSrc := filepath.Join(babelRoot, "rtl", version+".map")
		if info, err := os.Stat(mapSrc); err == nil && !info.IsDir() {
			mapDst := filepath.Join(gameDir, version+".map")
			if err := copyFile(mapSrc, mapDst); err != nil {
				return a.failIf(err)
			}
			a.logOK("Map copiee: " + mapDst)
		} else {
			a.log("Map non trouvee pour " + version + " (utiliser rlbabel-genmap si cette version n'est pas integree a la DLL).")
		}
	}

	if updateGameexe {
		gameexe := filepath.Join(gameDir, "GAMEEXE.INI")
		if err := updateBabelGameexe(gameexe, dllName, nameEnc); err != nil {
			return a.failIf(err)
		}
		a.logOK("GAMEEXE.INI mis a jour: " + gameexe)
	} else {
		a.log("GAMEEXE.INI laisse intact.")
	}

	if dllName == "rlBabelF.dll" {
		a.log("Note: rlBabelF sert aux vieux RealLive 1.2.x; il faut charger la DLL au demarrage avec LoadDLL(0, 'rlBabelF') ou via rlcInit().")
	} else {
		a.log("Note: pour RealLive 1.2.5+, GAMEEXE doit contenir une ligne #DLL.xxx = \"rlBabel\".")
	}
	a.logOK("Preparation Babel terminee.")
	return ""
}

func (a *App) RldevBabelWriteHeader(outputDir string, enableGlosses bool) string {
	a.log("========================================")
	a.log("  RLdev - Header Babel")
	a.log("========================================")

	if err := required("dossier de sortie", outputDir); err != nil {
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
	a.logOK("Header cree: " + path)
	a.log("Copie ces lignes au debut du script a tester, ou dans le header commun du projet.")
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
