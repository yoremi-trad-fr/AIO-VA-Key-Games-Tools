package main

import "testing"

func TestParseSSDumpOptionsDefaultsToUnfilteredExport(t *testing.T) {
	opts := parseSSDumpOptions(nil)
	if opts.ExportAllText || opts.DialogueOnly || opts.JapaneseOnly || opts.FullWidthOnly {
		t.Fatalf("unexpected default filters: %+v", opts)
	}
}

func TestParseSSDumpOptionsRecognizesDialogueOnlyAliases(t *testing.T) {
	for _, flag := range []string{"--dialogue-only", "--filter-tags"} {
		opts := parseSSDumpOptions([]string{flag})
		if !opts.DialogueOnly {
			t.Fatalf("%s did not enable DialogueOnly: %+v", flag, opts)
		}
	}
}

func TestParseSSDumpOptionsRecognizesJapaneseOnly(t *testing.T) {
	opts := parseSSDumpOptions([]string{"--japanese-only"})
	if !opts.JapaneseOnly {
		t.Fatalf("--japanese-only did not enable JapaneseOnly: %+v", opts)
	}
}
