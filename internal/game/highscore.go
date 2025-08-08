package game

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	configDirName = "pacman"
	highScoreFN   = "highscore.txt"
)

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
	return filepath.Join(dir, highScoreFN), nil
}

// LoadHighScore reads the persisted high score from disk.
// Returns 0 if the file is missing or cannot be parsed.
func LoadHighScore() int {
	path, err := highScoreFilePath()
	if err != nil {
		return 0
	}
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return 0
	}
	text := strings.TrimSpace(scanner.Text())
	n, err := strconv.Atoi(text)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// SaveHighScore writes the provided high score to disk.
// It is safe to call frequently; the file size is trivial.
func SaveHighScore(score int) error {
	if score < 0 {
		return errors.New("score must be non-negative")
	}
	path, err := highScoreFilePath()
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(score)), 0o644); err != nil {
		return err
	}
	// atomic-ish replace on most platforms
	return os.Rename(tmp, path)
}
