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

func TestGameKeyListIncludesSiglusTools061Entries(t *testing.T) {
	if len(GameKeys) < 80 {
		t.Fatalf("expected expanded SiglusTools 0.61 key list, got %d entries", len(GameKeys))
	}
}

func TestFindKeyNormalizesFullWidthNames(t *testing.T) {
	got, ok := FindKey("Rewrite +")
	if !ok {
		t.Fatal("expected Rewrite+ key")
	}
	if got.Key[0] != 0x36 || got.Key[15] != 0x94 {
		t.Fatalf("unexpected Rewrite+ key: %#v", got.Key)
	}
}

func TestFindKeyAcceptsUsefulAliases(t *testing.T) {
	got, ok := FindKey("Planetarian")
	if !ok {
		t.Fatal("expected Planetarian key")
	}
	if got.Name != "Planetarian HD Steam" {
		t.Fatalf("name = %q", got.Name)
	}

	got, ok = FindKey("Summer Pockets REFLECTION BLUE DL")
	if !ok {
		t.Fatal("expected Summer Pockets REFLECTION BLUE DL key")
	}
	if got.Key[0] != 0x08 || got.Key[15] != 0x4E {
		t.Fatalf("unexpected Summer Pockets REFLECTION BLUE DL key: %#v", got.Key)
	}
}

func TestFindKeyIncludesRecentKeyListNames(t *testing.T) {
	got, ok := FindKey("終のステラ DL")
	if !ok {
		t.Fatal("expected Tsui no Stella DL key")
	}
	if got.Key[0] != 0x4E || got.Key[15] != 0x1E {
		t.Fatalf("unexpected Tsui no Stella DL key: %#v", got.Key)
	}
}
