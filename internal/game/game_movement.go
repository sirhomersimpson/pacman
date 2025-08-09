package game

import (
	"math"
	"math/rand"

	"pacman/internal/entities"
)

func (g *Game) updatePlayerMovement() {
	// Attempt to turn when aligned to the center of a cell
	if g.isAlignedToCellCenter() && g.canMoveFromCellCenter(g.player.DesiredDir) {
		g.player.CurrentDir = g.player.DesiredDir
	}

	// Move in current direction if possible
	if g.canMove(g.player.CurrentDir) {
		dx, dy := entities.DirDelta(g.player.CurrentDir)
		g.player.X += float64(dx) * playerSpeedPixelsPerUpdate
		g.player.Y += float64(dy) * playerSpeedPixelsPerUpdate
	} else {
		// If blocked, snap to cell center to avoid jitter
		gx, gy := g.playerGrid()
		g.player.X = float64(gx*tileSize + tileSize/2)
		g.player.Y = float64(gy*tileSize + tileSize/2)
	}

	// Wrap-around tunnels
	maxX := float64(g.tileMap.Width * tileSize)
	if g.player.X < 0 {
		g.player.X += maxX
	}
	if g.player.X >= maxX {
		g.player.X -= maxX
	}
}

// Ghost behavior: random movement or fleeing behavior based on frightened state
func (g *Game) updateGhostsRandom() {
	for _, gh := range g.ghosts {
		// If not aligned to tile center, continue current direction
		gx := int(gh.X) / tileSize
		gy := int(gh.Y) / tileSize
		cx := float64(gx*tileSize + tileSize/2)
		cy := float64(gy*tileSize + tileSize/2)
		aligned := math.Abs(gh.X-cx) < 1.0 && math.Abs(gh.Y-cy) < 1.0

		if aligned {
			// Choose direction based on frightened state
			var chosenDir entities.Direction
			if g.isFrightened() {
				chosenDir = g.getFleeDirection(gh, gx, gy)
			} else {
				chosenDir = g.getRandomDirection(gh, gx, gy)
			}
			gh.CurrentDir = chosenDir
			// Snap to center when aligned to avoid drift
			gh.X = cx
			gh.Y = cy
		}

		// Move ghost with appropriate speed
		dx, dy := entities.DirDelta(gh.CurrentDir)
		speed := ghostSpeedPixelsPerUpdate
		if g.isFrightened() {
			speed *= 0.5 // 50% speed when frightened
		}

		// If blocked mid-cell, snap to center and pick a new direction immediately
		nextGX, nextGY := gx+dx, gy+dy
		if nextGX < 0 {
			nextGX = g.tileMap.Width - 1
		}
		if nextGX >= g.tileMap.Width {
			nextGX = 0
		}
		if g.tileMap.IsWall(nextGX, nextGY) && !aligned {
			gh.X = cx
			gh.Y = cy
			// force choose a new direction at this intersection
			aligned = true
		}

		gh.X += float64(dx) * speed
		gh.Y += float64(dy) * speed

		// wrap horizontally
		maxX := float64(g.tileMap.Width * tileSize)
		if gh.X < 0 {
			gh.X += maxX
		}
		if gh.X >= maxX {
			gh.X -= maxX
		}
		// clamp Y within bounds to avoid exiting map vertically
		minY := float64(tileSize / 2)
		maxY := float64(g.tileMap.Height*tileSize - tileSize/2)
		if gh.Y < minY {
			gh.Y = minY
		}
		if gh.Y > maxY {
			gh.Y = maxY
		}
	}
}

// getFleeDirection chooses the direction that maximizes distance from player
func (g *Game) getFleeDirection(gh *entities.Ghost, gx, gy int) entities.Direction {
	playerX, playerY := g.playerGrid()
	candidateDirs := []entities.Direction{entities.DirUp, entities.DirDown, entities.DirLeft, entities.DirRight}
	valid := make([]entities.Direction, 0, 4)

	// Find all valid directions
	for _, d := range candidateDirs {
		dx, dy := entities.DirDelta(d)
		nx, ny := gx+dx, gy+dy
		// Handle wrap-around
		if nx < 0 {
			nx = g.tileMap.Width - 1
		}
		if nx >= g.tileMap.Width {
			nx = 0
		}
		if ny < 0 || ny >= g.tileMap.Height {
			continue
		}
		if !g.tileMap.IsWall(nx, ny) {
			valid = append(valid, d)
		}
	}

	if len(valid) == 0 {
		// Emergency fallback
		return g.getRandomDirection(gh, gx, gy)
	}

	// Find direction that maximizes distance from player
	bestDir := valid[0]
	maxDistSq := float64(-1)

	for _, d := range valid {
		dx, dy := entities.DirDelta(d)
		nx, ny := gx+dx, gy+dy
		// Handle wrap-around for distance calculation
		if nx < 0 {
			nx = g.tileMap.Width - 1
		}
		if nx >= g.tileMap.Width {
			nx = 0
		}

		// Calculate distance squared to player
		distSq := float64((nx-playerX)*(nx-playerX) + (ny-playerY)*(ny-playerY))
		if distSq > maxDistSq {
			maxDistSq = distSq
			bestDir = d
		}
	}

	return bestDir
}

// getRandomDirection chooses a random valid direction (original behavior)
func (g *Game) getRandomDirection(gh *entities.Ghost, gx, gy int) entities.Direction {
	candidateDirs := []entities.Direction{entities.DirUp, entities.DirDown, entities.DirLeft, entities.DirRight}
	// place current direction first to bias straight
	ordered := make([]entities.Direction, 0, 4)
	if gh.CurrentDir != entities.DirNone {
		ordered = append(ordered, gh.CurrentDir)
	}
	for _, d := range candidateDirs {
		if d != gh.CurrentDir && !isReverse(gh.CurrentDir, d) {
			ordered = append(ordered, d)
		}
	}
	// If still empty (at start), just use candidates shuffled
	if len(ordered) == 0 {
		ordered = candidateDirs
	}
	rand.Shuffle(len(ordered), func(i, j int) { ordered[i], ordered[j] = ordered[j], ordered[i] })

	for _, d := range ordered {
		dx, dy := entities.DirDelta(d)
		nx, ny := gx+dx, gy+dy
		if nx < 0 {
			nx = g.tileMap.Width - 1
		}
		if nx >= g.tileMap.Width {
			nx = 0
		}
		if ny < 0 || ny >= g.tileMap.Height {
			continue
		}
		if !g.tileMap.IsWall(nx, ny) {
			return d
		}
	}

	// If no valid moves found, find any valid direction
	valid := make([]entities.Direction, 0, 4)
	for _, d := range candidateDirs {
		dx, dy := entities.DirDelta(d)
		nx, ny := gx+dx, gy+dy
		if nx < 0 {
			nx = g.tileMap.Width - 1
		}
		if nx >= g.tileMap.Width {
			nx = 0
		}
		if ny < 0 || ny >= g.tileMap.Height {
			continue
		}
		if !g.tileMap.IsWall(nx, ny) {
			valid = append(valid, d)
		}
	}

	if len(valid) == 0 {
		// Emergency fallback: try reverse direction first
		reverse := reverseDir(gh.CurrentDir)
		dx, dy := entities.DirDelta(reverse)
		nx, ny := gx+dx, gy+dy
		if nx < 0 {
			nx = g.tileMap.Width - 1
		}
		if nx >= g.tileMap.Width {
			nx = 0
		}
		if ny >= 0 && ny < g.tileMap.Height && !g.tileMap.IsWall(nx, ny) {
			return reverse
		}
		// Final fallback
		return entities.DirLeft
	}

	return valid[0]
}

func (g *Game) playerGrid() (int, int) {
	return int(g.player.X) / tileSize, int(g.player.Y) / tileSize
}

func (g *Game) cellCenter(gridX, gridY int) (float64, float64) {
	return float64(gridX*tileSize + tileSize/2), float64(gridY*tileSize + tileSize/2)
}

func (g *Game) isAlignedToCellCenter() bool {
	gx, gy := g.playerGrid()
	cx, cy := g.cellCenter(gx, gy)
	// Use alignment threshold to ensure we catch alignment at high speeds
	return math.Abs(g.player.X-cx) < alignmentThreshold && math.Abs(g.player.Y-cy) < alignmentThreshold
}

func (g *Game) isNearCellCenter() bool {
	gx, gy := g.playerGrid()
	cx, cy := g.cellCenter(gx, gy)
	return math.Abs(g.player.X-cx) < 5.0 && math.Abs(g.player.Y-cy) < 5.0
}

func (g *Game) canMove(dir entities.Direction) bool {
	if dir == entities.DirNone {
		return false
	}
	dx, dy := entities.DirDelta(dir)
	gx, gy := g.playerGrid()

	// If not aligned, only allow continuing straight to reach alignment
	if !g.isAlignedToCellCenter() && dir != g.player.CurrentDir {
		return false
	}

	// Next cell
	nx, ny := gx+dx, gy+dy
	// Wrap-around checks on X
	if nx < 0 {
		nx = g.tileMap.Width - 1
	}
	if nx >= g.tileMap.Width {
		nx = 0
	}
	if ny < 0 || ny >= g.tileMap.Height {
		return false
	}
	return !g.tileMap.IsWall(nx, ny)
}

// canMoveFromCellCenter checks if movement in a direction is valid from current cell
// without requiring perfect alignment (used for queued turns)
func (g *Game) canMoveFromCellCenter(dir entities.Direction) bool {
	if dir == entities.DirNone {
		return false
	}
	dx, dy := entities.DirDelta(dir)
	gx, gy := g.playerGrid()

	// Next cell
	nx, ny := gx+dx, gy+dy
	// Wrap-around checks on X
	if nx < 0 {
		nx = g.tileMap.Width - 1
	}
	if nx >= g.tileMap.Width {
		nx = 0
	}
	if ny < 0 || ny >= g.tileMap.Height {
		return false
	}
	return !g.tileMap.IsWall(nx, ny)
}

func isReverse(a, b entities.Direction) bool {
	return (a == entities.DirUp && b == entities.DirDown) ||
		(a == entities.DirDown && b == entities.DirUp) ||
		(a == entities.DirLeft && b == entities.DirRight) ||
		(a == entities.DirRight && b == entities.DirLeft)
}

func reverseDir(a entities.Direction) entities.Direction {
	switch a {
	case entities.DirUp:
		return entities.DirDown
	case entities.DirDown:
		return entities.DirUp
	case entities.DirLeft:
		return entities.DirRight
	case entities.DirRight:
		return entities.DirLeft
	default:
		return entities.DirLeft
	}
}
