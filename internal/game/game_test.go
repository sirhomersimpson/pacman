package game

import "testing"

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

