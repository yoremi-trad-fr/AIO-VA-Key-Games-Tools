package main

import (
	"reflect"
	"testing"
)

func TestSiglusSSDumpArgsAddsSingleLineForTXT(t *testing.T) {
	got := siglusSSDumpArgs(true, true, "smart", "txt", false)
	want := []string{"-d", "--single-line"}
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
