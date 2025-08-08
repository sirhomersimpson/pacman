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
