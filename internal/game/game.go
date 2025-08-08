package game

import (
	"fmt"
	"image/color"
	"math"

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
	playerSpeedPixelsPerSecond = 480.0
	playerSpeedPixelsPerUpdate = playerSpeedPixelsPerSecond / updatesPerSecond
)

type Game struct {
	tileMap    *tm.TileMap
	player     *entities.Player
	score      int
	fullscreen bool
	paused     bool
	quit       bool
	scale      float64
}

func New() *Game {
	m := tm.NewDefaultMap(tileSize)
	// Start player on a free corridor near bottom center (x=14, y=26 in default maze)
	startX := float64(14*tileSize + tileSize/2)
	startY := float64(26*tileSize + tileSize/2)
	p := &entities.Player{X: startX, Y: startY}
	g := &Game{tileMap: m, player: p}

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
	g.handleInput()
	if g.quit {
		return ebiten.Termination
	}
	if g.paused {
		return nil
	}
	g.updatePlayerMovement()
	g.handlePelletCollision()
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

	// HUD: Score
	text.Draw(off, fmt.Sprintf("Score: %d", g.score), basicfont.Face7x13, 4, 12, color.White)

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
			} else {
				g.score += 10
			}
		}
	}
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
