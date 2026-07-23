package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	if err := os.WriteFile(p, []byte("# comment\nFOO_KEY=abc123\nQUOTED=\"hello\"\n\nPRESET=fromfile\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PRESET", "fromenv") // real env must win
	t.Setenv("FOO_KEY", "")       // ensure unset-ish
	os.Unsetenv("FOO_KEY")
	os.Unsetenv("QUOTED")

	loadDotEnv(p)

	if os.Getenv("FOO_KEY") != "abc123" {
		t.Fatalf("FOO_KEY = %q, want abc123", os.Getenv("FOO_KEY"))
	}
	if os.Getenv("QUOTED") != "hello" {
		t.Fatalf("QUOTED = %q, want hello (quotes stripped)", os.Getenv("QUOTED"))
	}
	if os.Getenv("PRESET") != "fromenv" {
		t.Fatalf("PRESET = %q, want fromenv (real env should win)", os.Getenv("PRESET"))
	}
}
