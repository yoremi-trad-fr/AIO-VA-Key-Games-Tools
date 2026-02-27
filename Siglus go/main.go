package main

import (
	"fmt"
	"os"
	"siglustest/siglus"
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
			fmt.Println("usage: siglustest r <input_dir> <game_name> <wtfval_hex> <output.pck>")
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
		if err := siglus.RebuildPCK(os.Args[2], gk.Key, int32(wtfVal), os.Args[5]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	// ─── SS Dump (un fichier) ───────────────────────────────────
	case "dump":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest dump <file.ss> <output.tsv>")
			return
		}
		if err := siglus.DumpSSToTSV(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Dumped → %s\n", os.Args[3])
		}

	// ─── SS Dump (dossier entier) ───────────────────────────────
	case "dumpall":
		if len(os.Args) < 4 {
			fmt.Println("usage: siglustest dumpall <ss_dir> <tsv_output_dir>")
			return
		}
		if err := siglus.DumpSSDir(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

	// ─── SS Inject (un fichier) ─────────────────────────────────
	case "inject":
		if len(os.Args) < 5 {
			fmt.Println("usage: siglustest inject <original.ss> <translated.tsv> <output.ss>")
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
			fmt.Println("usage: siglustest injectall <original_ss_dir> <tsv_dir> <output_ss_dir>")
			return
		}
		if err := siglus.InjectSSDir(os.Args[2], os.Args[3], os.Args[4]); err != nil {
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

func printUsage() {
	fmt.Println("SiglusPCK tool - LuckSystem Yoremi fork")
	fmt.Println()
	fmt.Println("PCK operations:")
	fmt.Println("  x       <Scene.pck> <game> <output_dir>            Extract .ss files from PCK")
	fmt.Println("  r       <input_dir> <game> <wtfval> <output.pck>   Rebuild PCK from .ss files")
	fmt.Println()
	fmt.Println("SS text operations:")
	fmt.Println("  dump    <file.ss> <output.tsv>                     Dump text from one .ss")
	fmt.Println("  dumpall <ss_dir> <tsv_dir>                         Dump all .ss in a folder")
	fmt.Println("  inject  <orig.ss> <translated.tsv> <output.ss>     Inject translation into .ss")
	fmt.Println("  injectall <ss_dir> <tsv_dir> <output_dir>          Inject all TSV into .ss files")
	fmt.Println()
	fmt.Println("  keys                                                List available game keys")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  siglustest x Scene.pck Harmonia ./scene_extracted")
	fmt.Println("  siglustest dumpall ./scene_extracted ./text_fr")
	fmt.Println("  siglustest injectall ./scene_extracted ./text_fr ./scene_patched")
	fmt.Println("  siglustest r ./scene_patched Harmonia 0x566 Scene_patched.pck")
}
