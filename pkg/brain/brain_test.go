package brain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenMigratesLegacyPreSpectorDevSoul(t *testing.T) {
	dir := t.TempDir()
	legacy := "You are Macron, made by حاج متی (@Cianyd).\n"
	if err := os.WriteFile(filepath.Join(dir, SoulFile), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	newSoul := "You are Macron Agent, developed by SpectorDev Team (https://t.me/SpectorDev), MIT License."
	b, err := Open(dir, newSoul)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.SoulText(), "SpectorDev") || strings.Contains(b.SoulText(), "@Cianyd") {
		t.Fatalf("expected legacy SOUL.md to be migrated to SpectorDev, got: %q", b.SoulText())
	}
	if _, err := os.Stat(filepath.Join(dir, SoulFile+".pre-spectordev.bak")); err != nil {
		t.Fatalf("expected backup file to exist: %v", err)
	}
}

func TestOpenPreservesCustomSoul(t *testing.T) {
	dir := t.TempDir()
	custom := "Custom user persona — speak like a pirate.\n"
	if err := os.WriteFile(filepath.Join(dir, SoulFile), []byte(custom), 0o600); err != nil {
		t.Fatal(err)
	}
	b, err := Open(dir, "SpectorDev default")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(b.SoulText()) != strings.TrimSpace(custom) {
		t.Fatalf("custom SOUL.md was unexpectedly overwritten: %q", b.SoulText())
	}
}
