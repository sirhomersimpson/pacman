package tilemap

import "testing"

func TestNewDefaultMapDimensions(t *testing.T) {
	m := NewDefaultMap(16)
	if m.Width != len(defaultMaze[0]) || m.Height != len(defaultMaze) {
		t.Fatalf("unexpected dimensions: got %dx%d, want %dx%d", m.Width, m.Height, len(defaultMaze[0]), len(defaultMaze))
	}
}

func TestEatPelletAt(t *testing.T) {
	m := NewDefaultMap(16)
	var px, py int
	found := false
	for y := 0; y < m.Height && !found; y++ {
		for x := 0; x < m.Width && !found; x++ {
			if m.Tiles[y][x] == TilePellet {
				px, py = x, y
				found = true
			}
		}
	}
	if !found {
		t.Fatal("no pellet found in default map")
	}

	ate, power := m.EatPelletAt(px, py)
	if !ate || power {
		t.Fatalf("expected to eat normal pellet, got ate=%v power=%v", ate, power)
	}
	ate, power = m.EatPelletAt(px, py)
	if ate || power {
		t.Fatalf("expected to not eat after consumed, got ate=%v power=%v", ate, power)
	}
}

func TestIsWallBounds(t *testing.T) {
	m := NewDefaultMap(16)
	if !m.IsWall(-1, 0) || !m.IsWall(0, -1) || !m.IsWall(m.Width, 0) || !m.IsWall(0, m.Height) {
		t.Fatalf("out-of-bounds should be treated as wall")
	}
}
