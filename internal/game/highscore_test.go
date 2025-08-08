package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHighScoreLoadSaveWithEnvOverride(t *testing.T) {
	tdir := t.TempDir()
	if err := os.Setenv("PACMAN_CONFIG_DIR", tdir); err != nil {
		t.Fatalf("set env: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("PACMAN_CONFIG_DIR") })

	// Initially no file -> Load should be 0
	if got := LoadHighScore(); got != 0 {
		t.Fatalf("expected initial high score 0, got %d", got)
	}

	// Save a score and verify on disk
	if err := SaveHighScore(12345); err != nil {
		t.Fatalf("save high score: %v", err)
	}
	// Check file exists with content
	path, err := highScoreFilePath()
	if err != nil {
		t.Fatalf("highScoreFilePath: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "12345" {
		t.Fatalf("unexpected file contents: %q", string(data))
	}

	// Load should return the saved value
	if got := LoadHighScore(); got != 12345 {
		t.Fatalf("expected loaded 12345, got %d", got)
	}

	// Saving negative returns error and should not clobber file
	if err := SaveHighScore(-1); err == nil {
		t.Fatalf("expected error when saving negative score")
	}
	data2, _ := os.ReadFile(filepath.Join(tdir, "highscore.txt"))
	if string(data2) != "12345" {
		t.Fatalf("file should remain unchanged on error; got %q", string(data2))
	}
}
