package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"pacman/internal/entities"
	tm "pacman/internal/tilemap"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	tileSize                   = 16
	updatesPerSecond           = 60
	playerSpeedPixelsPerSecond = 720.0 // 480.0 * 1.5
	playerSpeedPixelsPerUpdate = playerSpeedPixelsPerSecond / updatesPerSecond
	ghostSpeedPixelsPerSecond  = 630.0 // 420.0 * 1.5
	ghostSpeedPixelsPerUpdate  = ghostSpeedPixelsPerSecond / updatesPerSecond
	frightenedDurationUpdates  = 120 // 120 ticks = 2 seconds at 60 UPS
)

type Game struct {
	tileMap             *tm.TileMap
	player              *entities.Player
	ghosts              []*entities.Ghost
	score               int
	lives               int
	fullscreen          bool
	paused              bool
	quit                bool
	scale               float64
	tickCounter         int
	frightenedUntilTick int
	ghostEatCombo       int
}

func New() *Game {
	rand.Seed(time.Now().UnixNano())
	m := tm.NewDefaultMap(tileSize)
	// Start player on a free corridor near bottom center (x=14, y=26 in default maze)
	startX := float64(14*tileSize + tileSize/2)
	startY := float64(26*tileSize + tileSize/2)
	p := &entities.Player{X: startX, Y: startY}
	g := &Game{tileMap: m, player: p, lives: 3}

	// Spawn 4 ghosts near the center (ghost house area) at nearest corridor tiles
	spawnTargets := [][2]int{{13, 14}, {14, 14}, {13, 15}, {14, 15}}
	for _, t := range spawnTargets {
		ox, oy := g.nearestCorridorTile(t[0], t[1])
		g.ghosts = append(g.ghosts, &entities.Ghost{
			X: float64(ox*tileSize + tileSize/2),
			Y: float64(oy*tileSize + tileSize/2),
		})
	}

	// Compute initial scale to fit within ~75% of the display area
	nativeW := m.Width * tileSize
	nativeH := m.Height * tileSize
	sw, sh := ebiten.ScreenSizeInFullscreen()
	fit := 0.75
	maxW := int(float64(sw) * fit)
	maxH := int(float64(sh) * fit)
	scaleW := float64(maxW) / float64(nativeW)
	scaleH := float64(maxH) / float64(nativeH)
	g.scale = math.Min(scaleW, scaleH)
	if g.scale <= 0 || math.IsNaN(g.scale) || math.IsInf(g.scale, 0) {
		g.scale = 1.0
	}
	return g
}

func (g *Game) ScreenWidth() int {
	return int(float64(g.tileMap.Width*tileSize) * g.scale)
}

func (g *Game) ScreenHeight() int {
	return int(float64(g.tileMap.Height*tileSize) * g.scale)
}

func (g *Game) Update() error {
	// Advance global tick counter first so timers are robust
	g.tickCounter++
	g.handleInput()
	if g.quit {
		return ebiten.Termination
	}

	if g.frightenedUntilTick != 0 && g.tickCounter >= g.frightenedUntilTick {
		g.frightenedUntilTick = 0
		g.ghostEatCombo = 0
	}
	if g.paused {
		return nil
	}
	g.updatePlayerMovement()
	g.handlePelletCollision()
	g.updateGhostsRandom()
	g.checkPlayerGhostCollision()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear background (black)
	screen.Fill(color.Black)

	// Use an offscreen image at native resolution then scale up
	nativeW := g.tileMap.Width * tileSize
	nativeH := g.tileMap.Height * tileSize
	off := ebiten.NewImage(nativeW, nativeH)

	// Draw map
	g.tileMap.Draw(off)

	// Draw player
	vector.DrawFilledCircle(off, float32(g.player.X), float32(g.player.Y), float32(tileSize/2-2), color.RGBA{R: 255, G: 221, B: 0, A: 255}, true)

	// Draw ghosts (simple circles)
	ghostColors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},     // red
		{R: 255, G: 128, B: 255, A: 255}, // pink
		{R: 255, G: 128, B: 0, A: 255},   // orange
		{R: 0, G: 191, B: 255, A: 255},   // cyan
	}
	for i, gh := range g.ghosts {
		c := ghostColors[i%len(ghostColors)]
		if g.isFrightened() {
			c = color.RGBA{R: 0, G: 0, B: 255, A: 255}
		}
		vector.DrawFilledCircle(off, float32(gh.X), float32(gh.Y), float32(tileSize/2-2), c, true)
	}

	// HUD: Score & Lives
	text.Draw(off, fmt.Sprintf("Score: %d  Lives: %d  FPS: %0.0f", g.score, g.lives, ebiten.ActualFPS()), basicfont.Face7x13, 4, 12, color.White)

	// Show frightened timer if active (bottom right corner)
	if g.isFrightened() {
		remainingTicks := g.frightenedUntilTick - g.tickCounter
		remainingSeconds := float64(remainingTicks) / float64(updatesPerSecond)
		timerText := fmt.Sprintf("Frightened: %.1fs", remainingSeconds)
		textWidth := len(timerText) * 7 // basicfont.Face7x13 is roughly 7 pixels wide per character
		text.Draw(off, timerText, basicfont.Face7x13, nativeW-textWidth-4, nativeH-4, color.RGBA{R: 0, G: 255, B: 255, A: 255})
	}

	// Scale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(g.scale, g.scale)
	screen.DrawImage(off, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ScreenWidth(), g.ScreenHeight()
}

func (g *Game) handleInput() {
	// Queue desired direction from input
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.player.DesiredDir = entities.DirUp
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.player.DesiredDir = entities.DirDown
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.player.DesiredDir = entities.DirLeft
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.player.DesiredDir = entities.DirRight
	}

	// Fullscreen toggle with 'F'
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}

	// Pause toggle with Space
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}

	// Quit with 'Q'
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.quit = true
	}
}

func (g *Game) updatePlayerMovement() {
	// Attempt to turn when aligned to the center of a cell
	if g.isAlignedToCellCenter() && g.canMove(g.player.DesiredDir) {
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

func (g *Game) handlePelletCollision() {
	// Eat pellet when close to cell center containing a pellet
	gx, gy := g.playerGrid()
	if g.isNearCellCenter() {
		ate, power := g.tileMap.EatPelletAt(gx, gy)
		if ate {
			if power {
				g.score += 50
				// Enter frightened mode for standard duration
				g.frightenedUntilTick = g.tickCounter + frightenedDurationUpdates
				g.ghostEatCombo = 0
			} else {
				g.score += 10
			}
		}
	}
}

// Ghost behavior: simple random movement on intersections, avoid walls
func (g *Game) updateGhostsRandom() {
	for _, gh := range g.ghosts {
		// If not aligned to tile center, continue current direction
		gx := int(gh.X) / tileSize
		gy := int(gh.Y) / tileSize
		cx := float64(gx*tileSize + tileSize/2)
		cy := float64(gy*tileSize + tileSize/2)
		aligned := math.Abs(gh.X-cx) < 1.0 && math.Abs(gh.Y-cy) < 1.0
		if aligned {
			// Choose a random valid direction, prefer continuing straight, avoid immediate reversal
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
					gh.CurrentDir = d
					break
				}
			}
			// Snap to center when aligned to avoid drift
			gh.X = cx
			gh.Y = cy
		}
		dx, dy := entities.DirDelta(gh.CurrentDir)
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
			// reuse aligned branch next iteration by marking aligned now
			aligned = true
		}
		if aligned {
			// Choose direction now that we are centered
			// replicate the same logic as above but simpler via recursion of branch
			// (inline):
			candidateDirs := []entities.Direction{entities.DirUp, entities.DirDown, entities.DirLeft, entities.DirRight}
			valid := make([]entities.Direction, 0, 4)
			for _, d := range candidateDirs {
				tx, ty := entities.DirDelta(d)
				nx, ny := gx+tx, gy+ty
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
				valid = []entities.Direction{reverseDir(gh.CurrentDir)}
			}
			rand.Shuffle(len(valid), func(i, j int) { valid[i], valid[j] = valid[j], valid[i] })
			// prefer straight if present
			for _, d := range valid {
				if d == gh.CurrentDir {
					gh.CurrentDir = d
					break
				}
			}
			gh.CurrentDir = valid[0]
			dx, dy = entities.DirDelta(gh.CurrentDir)
		}
		gh.X += float64(dx) * ghostSpeedPixelsPerUpdate
		gh.Y += float64(dy) * ghostSpeedPixelsPerUpdate
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

func (g *Game) checkPlayerGhostCollision() {
	// collision if within radius distance
	pr := float64(tileSize/2 - 2)
	gr := float64(tileSize/2 - 2)
	for _, gh := range g.ghosts {
		dx := g.player.X - gh.X
		dy := g.player.Y - gh.Y
		if dx*dx+dy*dy <= (pr+gr)*(pr+gr) {
			if g.isFrightened() {
				// Eat ghost: score increases with combo 200, 400, 800, 1600
				base := 200
				if g.ghostEatCombo > 0 {
					base = base << g.ghostEatCombo
				}
				if base > 1600 {
					base = 1600
				}
				g.score += base
				g.ghostEatCombo++
				// Send ghost back to house
				gh.X = float64(14*tileSize + tileSize/2)
				gh.Y = float64(14*tileSize + tileSize/2)
				gh.CurrentDir = entities.DirLeft
				continue
			}
			g.lives--
			g.resetPositions()
			return
		}
	}
}

func (g *Game) isFrightened() bool {
	return g.frightenedUntilTick > g.tickCounter
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

func (g *Game) playerGrid() (int, int) {
	return int(g.player.X) / tileSize, int(g.player.Y) / tileSize
}

func (g *Game) cellCenter(gridX, gridY int) (float64, float64) {
	return float64(gridX*tileSize + tileSize/2), float64(gridY*tileSize + tileSize/2)
}

func (g *Game) isAlignedToCellCenter() bool {
	gx, gy := g.playerGrid()
	cx, cy := g.cellCenter(gx, gy)
	return math.Abs(g.player.X-cx) < 1.0 && math.Abs(g.player.Y-cy) < 1.0
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
