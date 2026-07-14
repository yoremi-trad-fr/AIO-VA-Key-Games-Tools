package main

import (
	"reflect"
	"testing"
)

func TestSiglusSSDumpArgsAddsSingleLineForTXT(t *testing.T) {
	got := siglusSSDumpArgs(true, true, "all", "txt", false)
	want := []string{"-d", "--single-line", "-a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
}

func TestSiglusSSDumpArgsAddsDialogueOnlyFilter(t *testing.T) {
	got := siglusSSDumpArgs(false, true, "dialogue", "txt", false)
	want := []string{"--single-line", "--dialogue-only"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
}

func TestSiglusSSDumpArgsMapsLegacySmartFilterToDialogueOnly(t *testing.T) {
	got := siglusSSDumpArgs(false, false, "smart", "txt", false)
	want := []string{"--dialogue-only"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
}

func TestSiglusSSDumpArgsAddsJapaneseOnlyFilter(t *testing.T) {
	got := siglusSSDumpArgs(false, false, "japanese", "txt", false)
	want := []string{"--japanese-only"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
}

func TestSiglusSSDumpArgsIgnoresSingleLineForXLSX(t *testing.T) {
	got := siglusSSDumpArgs(true, true, "all", "xlsx", true)
	want := []string{"-d", "-a", "-x", "-s"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
}
