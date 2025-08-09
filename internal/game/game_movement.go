package game

import (
	"math"
	"math/rand"

	"pacman/internal/entities"
)

func (g *Game) updatePlayerMovement() {
	// Handle turning at intersections
	if g.player.DesiredDir != g.player.CurrentDir {
		// Check if we can turn
		if g.canTurn(g.player.DesiredDir) {
			// Snap to grid center on the perpendicular axis to avoid drift
			gx, gy := g.playerGrid()
			cx, cy := g.cellCenter(gx, gy)
			if g.player.DesiredDir == entities.DirUp || g.player.DesiredDir == entities.DirDown {
				g.player.X = cx
			} else if g.player.DesiredDir == entities.DirLeft || g.player.DesiredDir == entities.DirRight {
				g.player.Y = cy
			}
			g.player.CurrentDir = g.player.DesiredDir
		}
	}

	// Move in current direction
	if g.player.CurrentDir != entities.DirNone {
		dx, dy := entities.DirDelta(g.player.CurrentDir)
		newX := g.player.X + float64(dx)*playerSpeedPixelsPerUpdate
		newY := g.player.Y + float64(dy)*playerSpeedPixelsPerUpdate

		// Auto-center on the perpendicular axis when close to center to prevent drift
		gx, gy := g.playerGrid()
		cx, cy := g.cellCenter(gx, gy)
		if dx != 0 { // moving horizontally -> center Y
			// Only center Y if close; do not over-constrain
			if math.Abs(g.player.Y-cy) <= alignmentThreshold {
				newY = cy
			}
			if hardSnapEnabled {
				// Only snap X when we truly cross the exact center and are very close on Y
				nextX := newX
				if math.Abs(g.player.Y-cy) <= alignmentThreshold+hardSnapEpsilon {
					if (g.player.X-cx) > 0 && (nextX-cx) < 0 {
						newX = cx
					} else if (g.player.X-cx) < 0 && (nextX-cx) > 0 {
						newX = cx
					}
				}
			}
		} else if dy != 0 { // moving vertically -> center X
			if math.Abs(g.player.X-cx) <= alignmentThreshold {
				newX = cx
			}
			if hardSnapEnabled {
				nextY := newY
				if math.Abs(g.player.X-cx) <= alignmentThreshold+hardSnapEpsilon {
					if (g.player.Y-cy) > 0 && (nextY-cy) < 0 {
						newY = cy
					} else if (g.player.Y-cy) < 0 && (nextY-cy) > 0 {
						newY = cy
					}
				}
			}
		}

		// Check if the move is valid
		if g.isValidPosition(newX, newY) {
			g.player.X = newX
			g.player.Y = newY
		} else {
			// Stop if we hit a wall
			g.player.CurrentDir = entities.DirNone
		}
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

// canTurn checks if the player can turn in the desired direction
func (g *Game) canTurn(dir entities.Direction) bool {
	if dir == entities.DirNone {
		return false
	}

	// Get current grid position
	gx, gy := g.playerGrid()

	// Check if the target cell is walkable first
	dx, dy := entities.DirDelta(dir)
	nx, ny := gx+dx, gy+dy

	// Handle wrap-around
	if nx < 0 {
		nx = g.tileMap.Width - 1
	}
	if nx >= g.tileMap.Width {
		nx = 0
	}

	// Check bounds
	if ny < 0 || ny >= g.tileMap.Height {
		return false
	}

	// If target is a wall, can't turn
	if g.tileMap.IsWall(nx, ny) {
		return false
	}

	// For turning, require some alignment but add overshoot detection to be responsive
	cx, cy := g.cellCenter(gx, gy)

	// If requesting a vertical turn, check horizontal alignment or crossing of cell center
	if dir == entities.DirUp || dir == entities.DirDown {
		if math.Abs(g.player.X-cx) <= alignmentThreshold {
			return true
		}
		// If currently moving horizontally, detect if we'll cross the center next update
		cdx, _ := entities.DirDelta(g.player.CurrentDir)
		if cdx != 0 {
			nextX := g.player.X + float64(cdx)*playerSpeedPixelsPerUpdate
			// If the sign changes or we land exactly on center, allow the turn
			if (g.player.X-cx)*(nextX-cx) <= 0 {
				return true
			}
		}
		return false
	}

	// If requesting a horizontal turn, check vertical alignment or crossing of cell center
	if dir == entities.DirLeft || dir == entities.DirRight {
		if math.Abs(g.player.Y-cy) <= alignmentThreshold {
			return true
		}
		_, cdy := entities.DirDelta(g.player.CurrentDir)
		if cdy != 0 {
			nextY := g.player.Y + float64(cdy)*playerSpeedPixelsPerUpdate
			if (g.player.Y-cy)*(nextY-cy) <= 0 {
				return true
			}
		}
		return false
	}

	return true
}

// isValidPosition checks if a position is valid for the player
func (g *Game) isValidPosition(x, y float64) bool {
	// Check the center of the player
	gx := int(x) / tileSize
	gy := int(y) / tileSize

	// Handle wrap-around
	if gx < 0 {
		gx = g.tileMap.Width - 1
	}
	if gx >= g.tileMap.Width {
		gx = 0
	}

	// Check bounds
	if gy < 0 || gy >= g.tileMap.Height {
		return false
	}

	// Simple check: is the center cell a wall?
	if g.tileMap.IsWall(gx, gy) {
		return false
	}

	// Check if we're moving into a wall (edge detection)
	halfSize := float64(tileSize/2 - 3)

	// Check the four corners
	corners := [][2]int{
		{int(x-halfSize) / tileSize, int(y-halfSize) / tileSize},
		{int(x+halfSize) / tileSize, int(y-halfSize) / tileSize},
		{int(x-halfSize) / tileSize, int(y+halfSize) / tileSize},
		{int(x+halfSize) / tileSize, int(y+halfSize) / tileSize},
	}

	for _, corner := range corners {
		cx, cy := corner[0], corner[1]

		// Handle wrap
		if cx < 0 {
			cx = g.tileMap.Width - 1
		}
		if cx >= g.tileMap.Width {
			cx = 0
		}

		// Skip if out of bounds vertically
		if cy < 0 || cy >= g.tileMap.Height {
			return false
		}

		if g.tileMap.IsWall(cx, cy) {
			return false
		}
	}

	return true
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

		// Move ghost with appropriate speed, but only if not blocked
		speed := ghostSpeedPixelsPerUpdate
		if g.isFrightened() {
			speed *= 0.5 // 50% speed when frightened
		}

		// Use proper collision detection like player movement
		if g.canMoveGhost(gh, gh.CurrentDir) {
			dx, dy := entities.DirDelta(gh.CurrentDir)
			gh.X += float64(dx) * speed
			gh.Y += float64(dy) * speed
		} else {
			// If blocked, snap to center and force new direction choice
			gh.X = cx
			gh.Y = cy
			// Force alignment so ghost will choose new direction next frame
			aligned = true
			// Try to choose a new valid direction immediately
			var chosenDir entities.Direction
			if g.isFrightened() {
				chosenDir = g.getFleeDirection(gh, gx, gy)
			} else {
				chosenDir = g.getRandomDirection(gh, gx, gy)
			}
			gh.CurrentDir = chosenDir
		}

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

// canMoveGhost checks if a ghost can move in a direction from its current position
func (g *Game) canMoveGhost(gh *entities.Ghost, dir entities.Direction) bool {
	if dir == entities.DirNone {
		return false
	}
	dx, dy := entities.DirDelta(dir)
	gx := int(gh.X) / tileSize
	gy := int(gh.Y) / tileSize

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
