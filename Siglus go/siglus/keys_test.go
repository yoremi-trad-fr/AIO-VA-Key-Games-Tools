package siglus

import "testing"

func TestFindKeyAcceptsCommaSeparatedHex(t *testing.T) {
	got, ok := FindKey("0x7F, 0x0D, 0x88, 0x21, 0x7B, 0xEA, 0x41, 0xF3, 0xAA, 0x03, 0xA7, 0x2F, 0xEB, 0x60, 0xAD, 0x2E")
	if !ok {
		t.Fatal("expected hex key to parse")
	}
	if got.Name != "custom hex key" || got.Key[0] != 0x7F || got.Key[15] != 0x2E {
		t.Fatalf("unexpected parsed key: %#v", got)
	}
}

func TestFindKeyAcceptsCompactHex(t *testing.T) {
	got, ok := FindKey("7F0D88217BEA41F3AA03A72FEB60AD2E")
	if !ok {
		t.Fatal("expected compact hex key to parse")
	}
	if got.Key[1] != 0x0D || got.Key[14] != 0xAD {
		t.Fatalf("unexpected parsed key: %#v", got.Key)
	}
}

func TestFindKeyKeepsNameLookup(t *testing.T) {
	got, ok := FindKey("Harmonia")
	if !ok {
		t.Fatal("expected Harmonia key")
	}
	if got.Name != "Harmonia" {
		t.Fatalf("name = %q", got.Name)
	}
}
