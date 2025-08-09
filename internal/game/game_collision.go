package game

import "pacman/internal/entities"

func (g *Game) handlePelletCollision() {
	// Eat pellet when close to cell center containing a pellet
	gx, gy := g.playerGrid()
	if g.isNearCellCenter() {
		ate, power := g.tileMap.EatPelletAt(gx, gy)
		if ate {
			if power {
				g.score += powerPelletPoints
				// Enter frightened mode for standard duration
				g.frightenedUntilTick = g.tickCounter + frightenedDurationUpdates
				g.ghostEatCombo = 0
				// Reverse all ghosts when entering frightened mode
				g.reverseAllGhosts()
				if g.audio != nil {
					g.audio.PlayPowerPellet()
				}
			} else {
				g.score += pelletPoints
				if g.audio != nil {
					g.audio.PlayPellet()
				}
			}
			// Update and persist high score if surpassed
			if g.score > g.highScore {
				g.highScore = g.score
				_ = SaveHighScoreRecord(&HighScoreRecord{Name: g.playerName, Score: g.highScore})
			}
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
				base := baseGhostPoints
				if g.ghostEatCombo > 0 {
					base = base << g.ghostEatCombo
				}
				if base > maxGhostPoints {
					base = maxGhostPoints
				}
				g.score += base
				if g.audio != nil {
					g.audio.PlayGhostEaten()
				}
				if g.score > g.highScore {
					g.highScore = g.score
					_ = SaveHighScoreRecord(&HighScoreRecord{Name: g.playerName, Score: g.highScore})
				}
				g.ghostEatCombo++
				// Mark ghost as eaten: it should return to house quickly (eyes-only behavior)
				gh.State = entities.GhostEaten
				// Immediately choose a direction toward the house
				gh.CurrentDir = entities.DirNone
				continue
			}
			g.lives--
			if g.audio != nil {
				g.audio.PlayDeath()
			}
			g.resetPositions()
			if g.lives <= 0 {
				// Save best on game over
				if g.score > g.highScore {
					g.highScore = g.score
					_ = SaveHighScoreRecord(&HighScoreRecord{Name: g.playerName, Score: g.highScore})
				}
				// Show leaderboard instead of continuing
				g.showingLeaderboard = true
			}
			return
		}
	}
}
