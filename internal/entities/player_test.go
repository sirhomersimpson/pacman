package entities

import "testing"

func TestDirDelta(t *testing.T) {
	tests := []struct {
		name   string
		dir    Direction
		wantDX int
		wantDY int
	}{
		{name: "none", dir: DirNone, wantDX: 0, wantDY: 0},
		{name: "up", dir: DirUp, wantDX: 0, wantDY: -1},
		{name: "down", dir: DirDown, wantDX: 0, wantDY: 1},
		{name: "left", dir: DirLeft, wantDX: -1, wantDY: 0},
		{name: "right", dir: DirRight, wantDX: 1, wantDY: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dx, dy := DirDelta(tc.dir)
			if dx != tc.wantDX || dy != tc.wantDY {
				t.Fatalf("DirDelta(%v) = (%d,%d), want (%d,%d)", tc.dir, dx, dy, tc.wantDX, tc.wantDY)
			}
		})
	}
}
