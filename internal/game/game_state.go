package game

import (
	"pacman/internal/entities"
	tm "pacman/internal/tilemap"
)

func (g *Game) isFrightened() bool {
	return g.frightenedUntilTick > g.tickCounter
}

// reverseAllGhosts reverses the direction of all ghosts when entering frightened mode
func (g *Game) reverseAllGhosts() {
	for _, gh := range g.ghosts {
		gh.CurrentDir = reverseDir(gh.CurrentDir)
	}
}

func (g *Game) resetPositions() {
	// Reset player
	g.player.X = float64(14*tileSize + tileSize/2)
	g.player.Y = float64(26*tileSize + tileSize/2)
	g.player.CurrentDir = entities.DirNone
	g.player.DesiredDir = entities.DirNone
	// Clear frightened state on life loss
	g.frightenedUntilTick = 0
	g.ghostEatCombo = 0
	// Reset ghosts to house
	positions := [][2]int{{13, 14}, {14, 14}, {13, 15}, {14, 15}}
	for i, gh := range g.ghosts {
		ox, oy := g.nearestOpenTile(positions[i][0], positions[i][1])
		gh.X = float64(ox*tileSize + tileSize/2)
		gh.Y = float64(oy*tileSize + tileSize/2)
		gh.CurrentDir = entities.DirLeft
	}
}

// nearestOpenTile returns the nearest non-wall tile from a starting grid coordinate.
func (g *Game) nearestOpenTile(x, y int) (int, int) {
	if !g.tileMap.IsWall(x, y) {
		return x, y
	}
	// BFS ring search limited radius
	maxR := 6
	for r := 1; r <= maxR; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				nx, ny := x+dx, y+dy
				if nx < 0 || ny < 0 || nx >= g.tileMap.Width || ny >= g.tileMap.Height {
					continue
				}
				if !g.tileMap.IsWall(nx, ny) {
					return nx, ny
				}
			}
		}
	}
	// fallback to original
	return x, y
}

// nearestCorridorTile finds a non-wall, non-empty corridor (i.e., avoids large empty blue regions)
func (g *Game) nearestCorridorTile(x, y int) (int, int) {
	if !g.tileMap.IsWall(x, y) && g.tileMap.Tiles[y][x] != tm.TileEmpty {
		return x, y
	}
	maxR := 8
	for r := 1; r <= maxR; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				nx, ny := x+dx, y+dy
				if nx < 0 || ny < 0 || nx >= g.tileMap.Width || ny >= g.tileMap.Height {
					continue
				}
				if !g.tileMap.IsWall(nx, ny) && g.tileMap.Tiles[ny][nx] != tm.TileEmpty {
					return nx, ny
				}
			}
		}
	}
	return g.nearestOpenTile(x, y)
}
