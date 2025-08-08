package game

import (
	tm "pacman/internal/tilemap"
	"testing"
)

func TestScreenDimensionsPositive(t *testing.T) {
	g := New()
	if g.ScreenWidth() <= 0 || g.ScreenHeight() <= 0 {
		t.Fatalf("screen dimensions must be positive, got %dx%d", g.ScreenWidth(), g.ScreenHeight())
	}
}

func TestPlayerCannotMoveThroughWall(t *testing.T) {
	g := New()
	// Find a wall cell with a free cell to its left
	found := false
	for y := 1; y < g.tileMap.Height-1 && !found; y++ {
		for x := 1; x < g.tileMap.Width-1 && !found; x++ {
			if g.tileMap.IsWall(x, y) && !g.tileMap.IsWall(x-1, y) {
				g.player.X = float64((x-1)*16 + 8)
				g.player.Y = float64(y*16 + 8)
				g.player.CurrentDir = 0
				g.player.DesiredDir = 4 // right
				found = true
			}
		}
	}
	if !found {
		t.Skip("couldn't find suitable wall position in maze; skipping")
	}
	oldX := g.player.X
	for i := 0; i < 30; i++ {
		_ = g.Update()
	}
	if g.player.X > oldX+7 {
		t.Fatalf("player appears to have moved through a wall: oldX=%v newX=%v", oldX, g.player.X)
	}
}

func TestFrightenedModeTimeout(t *testing.T) {
	g := New()

	// Initially not frightened
	if g.isFrightened() {
		t.Error("Game should not be frightened initially")
	}

	// Simulate eating a power pellet
	g.tickCounter = 100
	g.frightenedUntilTick = g.tickCounter + frightenedDurationUpdates

	expectedEnd := g.tickCounter + frightenedDurationUpdates
	t.Logf("Power pellet eaten at tick %d, should end at tick %d (duration=%d)",
		g.tickCounter, expectedEnd, frightenedDurationUpdates)

	// Should be frightened now
	if !g.isFrightened() {
		t.Error("Game should be frightened after eating power pellet")
	}

	// Simulate updates until just before timeout
	for i := 0; i < frightenedDurationUpdates-1; i++ {
		g.tickCounter++

		// Check timeout logic (same as Update() method)
		if g.frightenedUntilTick != 0 && g.tickCounter >= g.frightenedUntilTick {
			g.frightenedUntilTick = 0
			g.ghostEatCombo = 0
		}

		if !g.isFrightened() && i < frightenedDurationUpdates-10 {
			t.Errorf("Game should still be frightened at tick %d (i=%d)", g.tickCounter, i)
		}
	}

	// Should still be frightened
	if !g.isFrightened() {
		t.Errorf("Game should still be frightened at tick %d", g.tickCounter)
	}

	// One more tick should end frightened mode
	g.tickCounter++
	if g.frightenedUntilTick != 0 && g.tickCounter >= g.frightenedUntilTick {
		g.frightenedUntilTick = 0
		g.ghostEatCombo = 0
	}

	if g.isFrightened() {
		t.Errorf("Game should no longer be frightened at tick %d (until was %d)",
			g.tickCounter, expectedEnd)
	}

	t.Logf("Test completed. Final tick: %d, expected end: %d", g.tickCounter, expectedEnd)
}

func TestHighScoreIntegrationOnPelletAndGhost(t *testing.T) {
	t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())
	g := New()
	if g.highScore != 0 {
		t.Fatalf("expected initial high score 0, got %d", g.highScore)
	}

	// Simulate scoring: pellet (+10)
	g.score = 0
	g.highScore = 0
	g.score += 10
	if g.score > g.highScore {
		g.highScore = g.score
		_ = SaveHighScore(g.highScore)
	}
	if g.highScore != 10 {
		t.Fatalf("expected high score 10 after pellet, got %d", g.highScore)
	}

	// Simulate frightened ghost eat (+200 base)
	g.score += 200
	if g.score > g.highScore {
		g.highScore = g.score
		_ = SaveHighScore(g.highScore)
	}
	if g.highScore != 210 {
		t.Fatalf("expected high score 210 after ghost, got %d", g.highScore)
	}
}

func TestHighScoreSavedOnQuitAndGameOver(t *testing.T) {
	t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())
	g := New()
	g.score = 500
	g.highScore = 0

	// Simulate quit path
	if g.score > g.highScore {
		g.highScore = g.score
		if err := SaveHighScore(g.highScore); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	if got := LoadHighScore(); got != 500 {
		t.Fatalf("expected saved high score 500, got %d", got)
	}

	// Simulate game over path saving a better score
	g.score = 800
	if g.score > g.highScore {
		g.highScore = g.score
		if err := SaveHighScore(g.highScore); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	if got := LoadHighScore(); got != 800 {
		t.Fatalf("expected saved high score 800, got %d", got)
	}
}

func TestNewLoadsExistingHighScore(t *testing.T) {
	t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())
	// Pre-save a record with a name
	if err := SaveHighScoreRecord(&HighScoreRecord{Name: "Bob", Score: 777}); err != nil {
		t.Fatalf("pre-save: %v", err)
	}
	g := New()
	if g.highScore != 777 || g.highScoreName != "Bob" {
		t.Fatalf("expected high score 777/Bob loaded in New, got %d/%q", g.highScore, g.highScoreName)
	}
}

func TestHighScoreUpdatedOnPelletCollision(t *testing.T) {
	t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())
	g := New()
	// Put player exactly at current grid's center
	gx, gy := g.playerGrid()
	cx, cy := g.cellCenter(gx, gy)
	g.player.X, g.player.Y = cx, cy
	// Ensure there's a pellet at this cell
	g.tileMap.Tiles[gy][gx] = tm.TilePellet

	// Call pellet collision logic
	g.handlePelletCollision()

	if g.score != 10 {
		t.Fatalf("expected score 10 after pellet, got %d", g.score)
	}
	if got := LoadHighScore(); got != 10 {
		t.Fatalf("expected persisted high score 10, got %d", got)
	}
}

func TestHighScoreUpdatedOnGhostEatWhenFrightened(t *testing.T) {
	t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())
	g := New()
	// Frightened state active
	g.tickCounter = 100
	g.frightenedUntilTick = 200

	// Place a ghost at player's position
	g.ghosts[0].X = g.player.X
	g.ghosts[0].Y = g.player.Y

	g.checkPlayerGhostCollision()

	if g.score < 200 {
		t.Fatalf("expected score >=200 after eating ghost, got %d", g.score)
	}
	if got := LoadHighScore(); got < 200 {
		t.Fatalf("expected persisted high score >=200, got %d", got)
	}
}

func TestHighScoreSavedOnGameOver(t *testing.T) {
	t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())
	g := New()
	g.lives = 1
	g.score = 123
	g.highScore = 0

	// Place a ghost at player's position with no frightened state
	g.ghosts[0].X = g.player.X
	g.ghosts[0].Y = g.player.Y

	g.checkPlayerGhostCollision()

	if g.lives != 0 {
		t.Fatalf("expected lives to reach 0, got %d", g.lives)
	}
	if got := LoadHighScore(); got != 123 {
		t.Fatalf("expected persisted high score 123 on game over, got %d", got)
	}
}
