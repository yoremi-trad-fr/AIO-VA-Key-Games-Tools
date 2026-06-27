package main

import (
	"fmt"
	"os"
	"path/filepath"
	"siglustest/siglus"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {

	// ─── PCK Extract ───────────────────────────────────────────
	case "x":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest x <Scene.pck> <game_name> <output_dir>")
			return
		}
		gk, ok := findKey(os.Args[3])
		if !ok {
			return
		}
		if err := siglus.ExtractPCK(os.Args[2], gk.Key, os.Args[4]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	// ─── PCK Rebuild ───────────────────────────────────────────
	case "r":
		if len(os.Args) < 6 {
			fmt.Println("usage: siglustest r <input_dir> <game_name> <wtfval_hex> <output.pck> [-c <2-17> | -f]")
			return
		}
		gk, ok := findKey(os.Args[3])
		if !ok {
			return
		}
		var wtfVal int64
		fmt.Sscanf(os.Args[4], "0x%x", &wtfVal)
		if wtfVal == 0 {
			fmt.Sscanf(os.Args[4], "%d", &wtfVal)
		}
		level, fake, _ := parsePackOptions(os.Args[6:])
		if err := siglus.RebuildPCKWithOptions(os.Args[2], gk.Key, int32(wtfVal), os.Args[5], level, fake); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	// ─── SS Dump (un fichier) ───────────────────────────────────
	case "dump":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest dump <file.ss> <output.txt|output.xlsx> [-d] [-a|-w] [-x]")
			return
		}
		opts := parseSSDumpOptions(os.Args[4:])
		var err error
		if shouldUseXLSX(os.Args[3], os.Args[4:]) {
			err = siglus.DumpSSToXLSX(os.Args[2], os.Args[3], opts)
		} else {
			err = siglus.DumpSSToText(os.Args[2], os.Args[3], opts)
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Dumped → %s\n", os.Args[3])
		}

	// ─── SS Dump (dossier entier) ───────────────────────────────
	case "dumpall":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest dumpall <ss_dir> <text_output_dir|output.xlsx> [-d] [-a|-w] [-x [-s]]")
			return
		}
		opts := parseSSDumpOptions(os.Args[4:])
		var err error
		if shouldUseXLSX(os.Args[3], os.Args[4:]) {
			if hasOption(os.Args[4:], "-s", "--single") || strings.EqualFold(filepath.Ext(os.Args[3]), ".xlsx") {
				err = siglus.DumpSSDirToSingleXLSX(os.Args[2], os.Args[3], opts)
			} else {
				err = siglus.DumpSSDirToXLSX(os.Args[2], os.Args[3], opts)
			}
		} else {
			err = siglus.DumpSSDirWithOptions(os.Args[2], os.Args[3], opts)
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	// ─── SS Inject (un fichier) ─────────────────────────────────
	case "inject":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest inject <original.ss> <translated.txt|translated.xlsx> <output.ss>")
			return
		}
		if err := siglus.InjectSS(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Injected → %s\n", os.Args[4])
		}

	// ─── SS Inject (dossier entier) ─────────────────────────────
	case "injectall":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest injectall <original_ss_dir> <text_or_xlsx_dir> <output_ss_dir>")
			return
		}
		if err := siglus.InjectSSDir(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	// ─── Gameexe.dat unpack ────────────────────────────────────
	case "gameexe-x":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest gameexe-x <Gameexe.dat> <game> <output.ini>")
			return
		}
		gk, ok := findKey(os.Args[3])
		if !ok {
			return
		}
		if err := siglus.UnpackGameexe(os.Args[2], gk.Key, os.Args[4]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Unpacked -> %s\n", os.Args[4])
		}

	// ─── Gameexe.dat pack ──────────────────────────────────────
	case "gameexe-r":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest gameexe-r <Gameexe.ini> <game> <output.dat> [-p] [-c <2-17> | -f]")
			return
		}
		gk, ok := findKey(os.Args[3])
		if !ok {
			return
		}
		level, fake, useKey := parsePackOptions(os.Args[5:])
		if err := siglus.PackGameexe(os.Args[2], gk.Key, useKey, level, fake, os.Args[4]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Packed -> %s\n", os.Args[4])
		}

	// ─── Mobile PCK unpack/repack ──────────────────────────────
	case "mpck-x":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest mpck-x <file.pck> <output_dir>")
			return
		}
		if err := siglus.UnpackMobilePCK(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Unpacked -> %s\n", os.Args[3])
		}

	case "mpck-r":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest mpck-r <input_dir> <output.pck>")
			return
		}
		if err := siglus.PackMobilePCK(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Packed -> %s\n", os.Args[3])
		}

	// ─── OMV header cut ────────────────────────────────────────
	case "omv-cut":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest omv-cut <input.omv> <output.ogv>")
			return
		}
		if err := siglus.CutOMVHeader(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Cut -> %s\n", os.Args[3])
		}

	case "omv-pack":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest omv-pack <input.ogv> <output.omv>")
			return
		}
		if err := siglus.PackOMV(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Packed -> %s\n", os.Args[3])
		}

	case "combine-png":
		if len(os.Args) < 3 {
			fmt.Println("usage: siglustest combine-png <png_dir> [output.png]")
			return
		}
		output := ""
		if len(os.Args) >= 4 {
			output = os.Args[3]
		}
		if err := siglus.CombinePNGDir(os.Args[2], output); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if output != "" {
			fmt.Printf("Combined -> %s\n", output)
		} else {
			fmt.Printf("Combined -> %s\n", siglus.DefaultCombinePNGOutput(os.Args[2]))
		}

	// ─── DBS dump/rebuild ──────────────────────────────────────
	case "dbs-x":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest dbs-x <file.dbs> <output.dbs.out> <output.dbs.txt> [-a] [-e shift-jis|gbk]")
			return
		}
		dumpAll := hasOption(os.Args[5:], "-a", "--all")
		encoding := parseEncodingOption(os.Args[5:], "shift-jis")
		if err := siglus.UnpackDBSWithEncoding(os.Args[2], os.Args[3], os.Args[4], dumpAll, encoding); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Unpacked -> %s / %s\n", os.Args[3], os.Args[4])
		}

	case "dbs-r":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest dbs-r <file.dbs.out> <file.dbs.txt> <output.dbs> [-c <2-17> | -f] [-e gbk|shift-jis]")
			return
		}
		level, fake, _ := parsePackOptions(os.Args[5:])
		encoding := parseEncodingOption(os.Args[5:], "gbk")
		if err := siglus.PackDBSWithOptions(os.Args[2], os.Args[3], os.Args[4], siglus.DBSPackOptions{
			CompressionLevel: level,
			FakeCompression:  fake,
			ANSIEncoding:     encoding,
		}); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Packed -> %s\n", os.Args[4])
		}

	case "dbs-xlsx":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest dbs-xlsx <file.dbs> <output.xlsx> [-e shift-jis|gbk]")
			return
		}
		encoding := parseEncodingOption(os.Args[4:], "shift-jis")
		if err := siglus.DumpDBSToXLSX(os.Args[2], os.Args[3], encoding); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Exported -> %s\n", os.Args[3])
		}

	case "dbs-build":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest dbs-build <xlsx_dir> <output_dir> [-e [gbk|shift-jis]] [-c <2-17> | -f]")
			return
		}
		if err := siglus.BuildDBSDirFromXLSX(os.Args[2], os.Args[3], parseDBSBuildOptions(os.Args[4:])); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	// ─── Liste des jeux ─────────────────────────────────────────
	case "keys":
		fmt.Println("Available games:")
		for _, name := range siglus.GameNameList() {
			fmt.Printf("  %s\n", name)
		}

	default:
		printUsage()
	}
}

func findKey(name string) (siglus.GameKey, bool) {
	gk, found := siglus.FindKey(name)
	if !found {
		fmt.Printf("Unknown game: %s\n", name)
		fmt.Println("Run 'siglustest keys' to list available games.")
		return siglus.GameKey{}, false
	}
	fmt.Printf("Using key for: %s\n", gk.Name)
	return gk, true
}

func parsePackOptions(args []string) (level int, fake bool, useKey bool) {
	level = 2
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-p", "--key":
			useKey = true
		case "-f", "--fake":
			fake = true
		case "-c", "--compression":
			level = 17
			if i+1 < len(args) {
				var parsed int
				if _, err := fmt.Sscanf(args[i+1], "%d", &parsed); err == nil {
					level = parsed
					i++
				}
			}
		}
	}
	return level, fake, useKey
}

func hasOption(args []string, names ...string) bool {
	for _, arg := range args {
		for _, name := range names {
			if arg == name {
				return true
			}
		}
	}
	return false
}

func parseEncodingOption(args []string, fallback string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-e" || args[i] == "--encoding" {
			return args[i+1]
		}
	}
	return fallback
}

func parseDBSBuildOptions(args []string) siglus.DBSBuildOptions {
	level, fake, _ := parsePackOptions(args)
	opts := siglus.DBSBuildOptions{
		CompressionLevel: level,
		FakeCompression:  fake,
		Unicode:          true,
		ANSIEncoding:     "gbk",
	}
	for i := 0; i < len(args); i++ {
		if args[i] != "-e" && args[i] != "--encoding" && args[i] != "--ansi" {
			continue
		}
		opts.Unicode = false
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			opts.ANSIEncoding = args[i+1]
		}
	}
	return opts
}

func shouldUseXLSX(outputPath string, args []string) bool {
	return hasOption(args, "-x", "--xlsx") || strings.EqualFold(filepath.Ext(outputPath), ".xlsx")
}

func parseSSDumpOptions(args []string) siglus.SSDumpOptions {
	return siglus.SSDumpOptions{
		CopyText:      hasOption(args, "-d", "--copy"),
		ExportAllText: hasOption(args, "-a", "--all"),
		FullWidthOnly: hasOption(args, "-w", "--full-width"),
	}
}

func printUsage() {
	fmt.Println("SiglusPCK tool - LuckSystem Yoremi fork")
	fmt.Println()
	fmt.Println("PCK operations:")
	fmt.Println("  x       <Scene.pck> <game> <output_dir>            Extract .ss files from PCK")
	fmt.Println("  r       <input_dir> <game> <wtfval> <output.pck>   Rebuild PCK from .ss files")
	fmt.Println()
	fmt.Println("SS text operations:")
	fmt.Println("  dump    <file.ss> <output.txt|xlsx>                Dump text from one .ss")
	fmt.Println("  dumpall <ss_dir> <text_dir|xlsx>                   Dump all .ss in a folder")
	fmt.Println("          options: -d copy, -a all, -w full-width, -x xlsx, -s single xlsx")
	fmt.Println("  inject  <orig.ss> <translated.txt|xlsx> <output.ss> Inject translation into .ss")
	fmt.Println("  injectall <ss_dir> <text_or_xlsx_dir> <output_dir> Inject all text into .ss files")
	fmt.Println()
	fmt.Println("Gameexe.dat operations:")
	fmt.Println("  gameexe-x <Gameexe.dat> <game> <output.ini>        Decrypt Gameexe.dat")
	fmt.Println("  gameexe-r <Gameexe.ini> <game> <output.dat>        Rebuild Gameexe.dat")
	fmt.Println()
	fmt.Println("DBS operations:")
	fmt.Println("  dbs-x   <file.dbs> <out.dbs.out> <out.dbs.txt>     Dump DBS data")
	fmt.Println("  dbs-r   <file.dbs.out> <file.dbs.txt> <out.dbs>    Rebuild DBS data")
	fmt.Println("  dbs-xlsx <file.dbs> <output.xlsx>                  Export DBS data to XLSX")
	fmt.Println("  dbs-build <xlsx_dir> <output_dir>                  Build DBS files from XLSX")
	fmt.Println()
	fmt.Println("Mobile PCK / OMV operations:")
	fmt.Println("  mpck-x  <file.pck> <output_dir>                    Extract mobile PCK")
	fmt.Println("  mpck-r  <input_dir> <output.pck>                   Rebuild mobile PCK")
	fmt.Println("  omv-cut <input.omv> <output.ogv>                   Remove OMV header")
	fmt.Println()
	fmt.Println("  keys                                                List available game keys")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  siglustest x Scene.pck Harmonia ./scene_extracted")
	fmt.Println("  siglustest dumpall ./scene_extracted ./text_fr")
	fmt.Println("  siglustest injectall ./scene_extracted ./text_fr ./scene_patched")
	fmt.Println("  siglustest r ./scene_patched Harmonia 0x166 Scene_patched.pck")
}
