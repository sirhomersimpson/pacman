package game

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	configDirName   = "pacman"
	highScoreTxtFN  = "highscore.txt"  // legacy
	highScoreJSONFN = "highscore.json" // current
)

// HighScoreRecord stores the top score and the name of the player who achieved it.
type HighScoreRecord struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// configBaseDir determines the base directory to store config.
// If PACMAN_CONFIG_DIR is set, it is used as-is. Otherwise, use UserConfigDir()/pacman.
func configBaseDir() (string, error) {
	if env := os.Getenv("PACMAN_CONFIG_DIR"); env != "" {
		if err := os.MkdirAll(env, 0o755); err != nil {
			return "", err
		}
		return env, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, configDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// highScoreFilePath returns the absolute path to the high score file.
func highScoreFilePath() (string, error) {
	dir, err := configBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, highScoreJSONFN), nil
}

// LoadHighScore reads the persisted high score (score only). Prefer JSON, fallback to legacy txt.
// Returns 0 if nothing is found.
func LoadHighScore() int {
	if rec := LoadHighScoreRecord(); rec != nil {
		return rec.Score
	}
	return 0
}

// SaveHighScore writes only the score to disk, using an empty name if none known.
// Prefer using SaveHighScoreRecord for new code.
func SaveHighScore(score int) error {
	if score < 0 {
		return errors.New("score must be non-negative")
	}
	current := LoadHighScoreRecord()
	name := ""
	if current != nil {
		name = current.Name
	}
	return SaveHighScoreRecord(&HighScoreRecord{Name: name, Score: score})
}

// LoadHighScoreRecord returns the highest-score record from the leaderboard.
func LoadHighScoreRecord() *HighScoreRecord {
	list := LoadLeaderboard()
	if len(list) == 0 {
		return nil
	}
	// list is not guaranteed sorted
	best := list[0]
	for _, r := range list[1:] {
		if r.Score > best.Score {
			best = r
		}
	}
	return &best
}

// SaveHighScoreRecord upserts the provided record into the leaderboard and writes JSON array atomically.
func SaveHighScoreRecord(rec *HighScoreRecord) error {
	if rec == nil {
		return errors.New("nil record")
	}
	if rec.Score < 0 {
		return errors.New("score must be non-negative")
	}
	dir, err := configBaseDir()
	if err != nil {
		return err
	}
	// Load existing leaderboard (object or array)
	leaderboard := LoadLeaderboard()
	updated := false
	for i := range leaderboard {
		if strings.EqualFold(strings.TrimSpace(leaderboard[i].Name), strings.TrimSpace(rec.Name)) {
			if rec.Score > leaderboard[i].Score {
				leaderboard[i].Score = rec.Score
			}
			updated = true
			break
		}
	}
	if !updated {
		leaderboard = append(leaderboard, *rec)
	}
	// Write as JSON array
	path := filepath.Join(dir, highScoreJSONFN)
	tmp := path + ".tmp"
	data, err := json.MarshalIndent(leaderboard, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadLeaderboard loads all known high score records.
// Accepts either a JSON array of records (preferred) or a single record object for backward compatibility.
// Falls back to legacy txt format if present.
func LoadLeaderboard() []HighScoreRecord {
	dir, err := configBaseDir()
	if err != nil {
		return nil
	}
	jpath := filepath.Join(dir, highScoreJSONFN)
	if data, err := os.ReadFile(jpath); err == nil {
		// Try array first
		var arr []HighScoreRecord
		if err := json.Unmarshal(data, &arr); err == nil {
			return arr
		}
		// Try single object
		var obj HighScoreRecord
		if err := json.Unmarshal(data, &obj); err == nil && obj.Score >= 0 {
			return []HighScoreRecord{obj}
		}
	}
	// Fallback to legacy txt
	tpath := filepath.Join(dir, highScoreTxtFN)
	if f, err := os.Open(tpath); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			n, err := strconv.Atoi(text)
			if err == nil && n >= 0 {
				return []HighScoreRecord{{Name: "", Score: n}}
			}
		}
	}
	return nil
}
