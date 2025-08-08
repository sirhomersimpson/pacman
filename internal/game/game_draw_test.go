package game

import (
	"testing"

	tm "pacman/internal/tilemap"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestGameDrawDoesNotPanic(t *testing.T) {
	t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())
	g := New()
	screen := ebiten.NewImage(g.ScreenWidth(), g.ScreenHeight())
	// Should not panic
	g.Draw(screen)
}

func TestLayoutMatchesScreenSize(t *testing.T) {
	g := New()
	w, h := g.Layout(0, 0)
	if w != g.ScreenWidth() || h != g.ScreenHeight() {
		t.Fatalf("layout mismatch: got %dx%d want %dx%d", w, h, g.ScreenWidth(), g.ScreenHeight())
	}
}

func TestReverseDirMapping(t *testing.T) {
	if reverseDir(0) == 0 { // DirNone -> should return a valid dir (Left)
		t.Fatalf("reverseDir for none should not be none")
	}
}

func TestNearestOpenTileVariants(t *testing.T) {
	g := New()

	// Case 1: Already open tile returns itself
	openX, openY := -1, -1
	for y := 0; y < g.tileMap.Height && openX < 0; y++ {
		for x := 0; x < g.tileMap.Width && openX < 0; x++ {
			if !g.tileMap.IsWall(x, y) {
				openX, openY = x, y
			}
		}
	}
	if openX < 0 {
		t.Skip("no open tile found")
	}
	nx, ny := g.nearestOpenTile(openX, openY)
	if nx != openX || ny != openY {
		t.Fatalf("expected same coords for open tile, got %d,%d vs %d,%d", nx, ny, openX, openY)
	}

	// Case 2: Wall tile returns a nearby open tile
	wallX, wallY := -1, -1
	for y := 0; y < g.tileMap.Height && wallX < 0; y++ {
		for x := 0; x < g.tileMap.Width && wallX < 0; x++ {
			if g.tileMap.IsWall(x, y) {
				wallX, wallY = x, y
			}
		}
	}
	if wallX < 0 {
		t.Skip("no wall tile found")
	}
	nx, ny = g.nearestOpenTile(wallX, wallY)
	if g.tileMap.IsWall(nx, ny) {
		t.Fatalf("expected non-wall from nearestOpenTile, got wall at %d,%d", nx, ny)
	}

	// Case 3: Fallback when all walls
	// Force all walls
	for y := 0; y < g.tileMap.Height; y++ {
		for x := 0; x < g.tileMap.Width; x++ {
			g.tileMap.Tiles[y][x] = tm.TileWall
		}
	}
	fx, fy := g.nearestOpenTile(5, 5)
	if fx != 5 || fy != 5 {
		t.Fatalf("expected fallback to original when all walls, got %d,%d", fx, fy)
	}
}

func TestCanMoveWhenNotAligned(t *testing.T) {
	g := New()
	// Put player between centers to ensure not aligned
	g.player.X += 2.0
	g.player.Y += 2.0
	g.player.CurrentDir = 0
	// Attempt to move Right while not aligned should be false
	if g.canMove(4) { // DirRight
		t.Fatalf("expected canMove false when not aligned and turning")
	}
}
