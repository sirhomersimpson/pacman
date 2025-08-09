package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strings"
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

	// Alignment and movement constants
	alignmentThreshold = playerSpeedPixelsPerUpdate / 2.0 // 6 pixels for responsive turning

	// Easter egg constants
	easterEggChance     = 6000 // 1 in 6000 chance (~100 seconds at 60 UPS)
	easterEggDuration   = 3    // 3 seconds
	maxPlayerNameLength = 12

	// Scoring constants
	pelletPoints      = 10
	powerPelletPoints = 50
	baseGhostPoints   = 200
	maxGhostPoints    = 1600

	// Display constants
	displayFitRatio = 0.75 // Use 75% of display area
	fontCharWidth   = 7    // basicfont.Face7x13 character width
)

type Game struct {
	tileMap             *tm.TileMap
	player              *entities.Player
	ghosts              []*entities.Ghost
	score               int
	highScore           int
	highScoreName       string
	playerName          string
	enteringName        bool
	showingLeaderboard  bool
	lives               int
	fullscreen          bool
	paused              bool
	quit                bool
	scale               float64
	tickCounter         int
	frightenedUntilTick int
	ghostEatCombo       int
	audio               *AudioManager
	easterMessage       string
	easterUntilTick     int
	offscreenImage      *ebiten.Image // Cached to avoid per-frame allocation
}

func New() *Game {
	rand.Seed(time.Now().UnixNano())
	m := tm.NewDefaultMap(tileSize)
	// Start player on a free corridor near bottom center (x=14, y=26 in default maze)
	startX := float64(14*tileSize + tileSize/2)
	startY := float64(26*tileSize + tileSize/2)
	p := &entities.Player{X: startX, Y: startY}
	g := &Game{tileMap: m, player: p, lives: 3}

	// Load persisted high score (with name if present)
	if rec := LoadHighScoreRecord(); rec != nil {
		g.highScore = rec.Score
		g.highScoreName = rec.Name
	} else {
		g.highScore = 0
	}
	g.enteringName = true

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
	maxW := int(float64(sw) * displayFitRatio)
	maxH := int(float64(sh) * displayFitRatio)
	scaleW := float64(maxW) / float64(nativeW)
	scaleH := float64(maxH) / float64(nativeH)
	g.scale = math.Min(scaleW, scaleH)
	if g.scale <= 0 || math.IsNaN(g.scale) || math.IsInf(g.scale, 0) {
		g.scale = 1.0
	}

	// Initialize cached offscreen image
	g.offscreenImage = ebiten.NewImage(nativeW, nativeH)

	// Init audio (graceful if files missing)
	g.audio = NewAudioManager("assets/sounds")
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

	// Clear easter egg message when time elapses
	if g.easterUntilTick != 0 && g.tickCounter >= g.easterUntilTick {
		g.easterUntilTick = 0
		g.easterMessage = ""
	}

	// Random, rare easter-egg trigger (about ~100s on average)
	if !g.enteringName && !g.showingLeaderboard && g.easterMessage == "" {
		// Roughly 1 in 6000 updates (~100 seconds at 60 UPS)
		if rand.Intn(easterEggChance) == 0 {
			if rand.Intn(2) == 0 {
				g.easterMessage = "Dad Loves Rekha"
			} else {
				g.easterMessage = "Dad Loves Roy"
			}
			g.easterUntilTick = g.tickCounter + updatesPerSecond*easterEggDuration
		}
	}

	if g.showingLeaderboard {
		return nil
	}

	if g.enteringName {
		return nil
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

	// Use cached offscreen image at native resolution then scale up
	off := g.offscreenImage
	off.Fill(color.Black) // Clear the cached image

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
			remainingTicks := g.frightenedUntilTick - g.tickCounter
			// Flash white/blue in last 2 seconds (120 ticks)
			if remainingTicks < 120 {
				// Alternate between white and blue every 10 ticks
				if (g.tickCounter/10)%2 == 0 {
					c = color.RGBA{R: 255, G: 255, B: 255, A: 255} // white
				} else {
					c = color.RGBA{R: 0, G: 0, B: 255, A: 255} // blue
				}
			} else {
				// Solid blue when not flashing
				c = color.RGBA{R: 0, G: 0, B: 255, A: 255}
			}
		}
		vector.DrawFilledCircle(off, float32(gh.X), float32(gh.Y), float32(tileSize/2-2), c, true)
	}

	// HUD: Score, High Score (with name) & Lives
	hiLabel := "High"
	if g.highScoreName != "" {
		hiLabel = fmt.Sprintf("High(%s)", g.highScoreName)
	}
	name := g.playerName
	if name == "" {
		name = "Player"
	}
	text.Draw(off, fmt.Sprintf("%s  Score: %d  %s: %d  Lives: %d  FPS: %0.0f", name, g.score, hiLabel, g.highScore, g.lives, ebiten.ActualFPS()), basicfont.Face7x13, 4, 12, color.White)

	// Show frightened timer if active (bottom right corner)
	if g.isFrightened() {
		remainingTicks := g.frightenedUntilTick - g.tickCounter
		remainingSeconds := float64(remainingTicks) / float64(updatesPerSecond)
		timerText := fmt.Sprintf("Frightened: %.1fs", remainingSeconds)
		textWidth := len(timerText) * fontCharWidth
		nativeW := g.tileMap.Width * tileSize
		nativeH := g.tileMap.Height * tileSize
		text.Draw(off, timerText, basicfont.Face7x13, nativeW-textWidth-4, nativeH-4, color.RGBA{R: 0, G: 255, B: 255, A: 255})
	}

	// If awaiting name, draw prompt centered
	if g.enteringName {
		prompt := "Enter name: " + g.playerName + "_"
		pw := len(prompt) * fontCharWidth
		nativeW := g.tileMap.Width * tileSize
		nativeH := g.tileMap.Height * tileSize
		text.Draw(off, prompt, basicfont.Face7x13, (nativeW-pw)/2, nativeH/2, color.White)
	}

	// If showing leaderboard, draw it centered
	if g.showingLeaderboard {
		list := LoadLeaderboard()
		nativeW := g.tileMap.Width * tileSize
		nativeH := g.tileMap.Height * tileSize
		title := "High Scores"
		tw := len(title) * fontCharWidth
		y := nativeH/2 - 40
		text.Draw(off, title, basicfont.Face7x13, (nativeW-tw)/2, y, color.RGBA{R: 255, G: 215, B: 0, A: 255})
		y += 14

		// Sort by score descending using efficient sort.Slice
		sort.Slice(list, func(i, j int) bool {
			return list[i].Score > list[j].Score
		})

		// Limit to top 10
		displayCount := len(list)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			line := fmt.Sprintf("%2d. %-12s  %6d", i+1, list[i].Name, list[i].Score)
			lw := len(line) * fontCharWidth
			text.Draw(off, line, basicfont.Face7x13, (nativeW-lw)/2, y, color.White)
			y += 14
		}
		hint := "Press Q to exit"
		hw := len(hint) * fontCharWidth
		text.Draw(off, hint, basicfont.Face7x13, (nativeW-hw)/2, nativeH-8, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	}

	// Draw easter egg message if present (overlay)
	if g.easterMessage != "" {
		msg := g.easterMessage
		mw := len(msg) * fontCharWidth
		nativeW := g.tileMap.Width * tileSize
		nativeH := g.tileMap.Height * tileSize
		text.Draw(off, msg, basicfont.Face7x13, (nativeW-mw)/2, nativeH/2-20, color.RGBA{R: 255, G: 192, B: 203, A: 255})
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
	// Name entry handling takes precedence
	if g.enteringName {
		// Collect typed characters
		var chars []rune
		chars = ebiten.AppendInputChars(chars)
		for _, r := range chars {
			if len([]rune(g.playerName)) >= maxPlayerNameLength {
				break
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '_' || r == '-' {
				g.playerName += string(r)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			rs := []rune(g.playerName)
			if len(rs) > 0 {
				g.playerName = string(rs[:len(rs)-1])
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
			if len([]rune(g.playerName)) > 0 {
				g.enteringName = false
				// Name-based easter eggs
				low := strings.ToLower(strings.TrimSpace(g.playerName))
				if low == "rekha" {
					g.easterMessage = "Dad Loves Rekha"
					g.easterUntilTick = g.tickCounter + updatesPerSecond*easterEggDuration
				}
				if low == "roy" {
					g.easterMessage = "Dad Loves Roy"
					g.easterUntilTick = g.tickCounter + updatesPerSecond*easterEggDuration
				}
			}
		}
		// Allow quitting/fullscreen even while entering name
		if inpututil.IsKeyJustPressed(ebiten.KeyF) {
			g.fullscreen = !g.fullscreen
			ebiten.SetFullscreen(g.fullscreen)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
			// Save if necessary with name
			if g.score > g.highScore {
				g.highScore = g.score
				_ = SaveHighScoreRecord(&HighScoreRecord{Name: g.playerName, Score: g.highScore})
			}
			g.quit = true
		}
		return
	}

	// Don't process movement input when game is not actively playing
	if g.showingLeaderboard || g.paused {
		// Skip movement input but still allow other keys
	} else {
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
		// Persist high score before quitting
		if g.score > g.highScore {
			g.highScore = g.score
			_ = SaveHighScoreRecord(&HighScoreRecord{Name: g.playerName, Score: g.highScore})
		}
		// If leaderboard showing already, exit; otherwise show it first
		if g.showingLeaderboard {
			g.quit = true
		} else {
			g.showingLeaderboard = true
		}
	}

	// Show leaderboard with 'S'
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.showingLeaderboard = !g.showingLeaderboard
	}

	// Easter eggs via keys
	if inpututil.IsKeyJustPressed(ebiten.KeyR) { // Rekha
		g.easterMessage = "Dad Loves Rekha"
		g.easterUntilTick = g.tickCounter + updatesPerSecond*easterEggDuration
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyY) { // Roy
		g.easterMessage = "Dad Loves Roy"
		g.easterUntilTick = g.tickCounter + updatesPerSecond*easterEggDuration
	}
}
