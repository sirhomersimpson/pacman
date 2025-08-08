package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHighScoreJSONRecordLoadSave(t *testing.T) {
	tdir := t.TempDir()
	t.Setenv("PACMAN_CONFIG_DIR", tdir)

	// No record yet
	if rec := LoadHighScoreRecord(); rec != nil {
		t.Fatalf("expected no record, got %+v", rec)
	}

	// Save record
	in := &HighScoreRecord{Name: "Alice", Score: 321}
	if err := SaveHighScoreRecord(in); err != nil {
		t.Fatalf("save record: %v", err)
	}

	// Load and compare
	out := LoadHighScoreRecord()
	if out == nil || out.Name != "Alice" || out.Score != 321 {
		t.Fatalf("unexpected record: %+v", out)
	}

	// Save score-only path preserves name
	if err := SaveHighScore(456); err != nil {
		t.Fatalf("save score: %v", err)
	}
	out2 := LoadHighScoreRecord()
	if out2 == nil || out2.Name != "Alice" || out2.Score != 456 {
		t.Fatalf("unexpected record after score save: %+v", out2)
	}
}

func TestHighScoreLegacyFallback(t *testing.T) {
	tdir := t.TempDir()
	t.Setenv("PACMAN_CONFIG_DIR", tdir)
	// Write legacy txt
	if err := os.WriteFile(filepath.Join(tdir, "highscore.txt"), []byte("999"), 0o644); err != nil {
		t.Fatalf("write legacy: %v", err)
	}
	rec := LoadHighScoreRecord()
	if rec == nil || rec.Score != 999 || rec.Name != "" {
		t.Fatalf("expected legacy load score=999 name='', got %+v", rec)
	}
}

func TestSaveRecordUpsertAndSort(t *testing.T) {
	tdir := t.TempDir()
	t.Setenv("PACMAN_CONFIG_DIR", tdir)
	// Add two players
	_ = SaveHighScoreRecord(&HighScoreRecord{Name: "Ana", Score: 100})
	_ = SaveHighScoreRecord(&HighScoreRecord{Name: "Bob", Score: 200})
	// Update Ana to higher than Bob
	_ = SaveHighScoreRecord(&HighScoreRecord{Name: "Ana", Score: 300})
	list := LoadLeaderboard()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	// ensure Ana has 300
	var anaScore, bobScore int
	for _, r := range list {
		if r.Name == "Ana" {
			anaScore = r.Score
		}
		if r.Name == "Bob" {
			bobScore = r.Score
		}
	}
	if anaScore != 300 || bobScore != 200 {
		t.Fatalf("unexpected scores: Ana=%d Bob=%d", anaScore, bobScore)
	}
}
