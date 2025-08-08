Pac-Man (Go + Ebiten)

Overview

This is a Pac-Man style game written in Go using Ebiten. It features a 28×31 ASCII maze, grid-based movement, pellets and power pellets, four ghosts with random movement (Milestone 3), frightened mode, scoring, named high scores with a persistent leaderboard, and simple audio cues (disabled by default).

Quick Start

- Prereqs: Go 1.19+ (tested) / 1.22+ recommended
- Build & run:
  - macOS/Linux: `make run`
  - Windows: `make build` then run the binary in `bin/`

Controls

- Arrow keys: Move
- Space: Pause
- F: Toggle fullscreen
- Q: Quit (if leaderboard not showing, it will show first; press Q again to exit)
- S: Toggle leaderboard view
- R: Show “Dad Loves Rekha” (easter egg)
- Y: Show “Dad Loves Roy” (easter egg)

Name & High Scores

- On first screen, type your name and press Enter.
- Your high scores persist to `highscore.json` under your OS user config directory.
- A leaderboard (top entries by score) is displayed on game over, or by pressing S.
- The leaderboard is backward compatible: it can import an old `highscore.txt` the next time you save.

Audio

- Audio is disabled by default.
- Enable audio per run:
  - macOS/Linux: `PACMAN_ENABLE_AUDIO=1 make run`
  - Windows (PowerShell): `$env:PACMAN_ENABLE_AUDIO='1'; make run`
- To force-disable: set `PACMAN_DISABLE_AUDIO=1` (takes precedence).
- Place WAV files in `assets/sounds/` named:
  - `pellet.wav`, `power.wav`, `ghost.wav`, `death.wav`
- If assets are missing, the game falls back to simple synthesized beeps.

Development

- Format & vet: `make fmt` and `make vet`
- Tests: `make test`
- Coverage:
  - Summary: `make coverage`
  - HTML report: `make coverage-html` then open `coverage.html`

Project Structure

- `cmd/pacman`: entry point
- `internal/game`: core game loop, state, UI, audio, leaderboard
- `internal/entities`: player and ghost definitions
- `internal/tilemap`: maze, tile map rendering and utility
- `internal/ui`: HUD placeholder utilities
- `assets/`: images and sounds

Notable Features

- Fixed timestep (60 UPS), grid-aligned movement
- Frightened mode timer and ghost-eating combo (200→1600)
- Named high score persistence with leaderboard UI
- Leaderboard shown on game over or before quitting
- Rare/random Easter eggs and name-based Easter eggs

License

See `LICENSE`.


