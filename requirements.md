Here’s a **`requirements.md`** you can drop straight into Cursor or Claude Code to kickstart your Go/Ebiten Pac-Man project.

---

# Pac-Man Clone — Requirements (Go + Ebiten)

## 1. Overview

We are building a Pac-Man style game in Go using the [Ebiten](https://ebiten.org/) game library.
The game should closely mimic the classic Pac-Man look and feel (grid layout, pellets, ghosts, scoring, power-ups) but with **original, non-copyrighted assets**.

---

## 2. Goals

* **Cross-platform** (macOS, Linux, Windows).
* **Single executable build**.
* **Maintainable code** split into logical packages.
* **Deterministic gameplay** (fixed time step, grid-aligned movement).
* **Easily replaceable art/sound assets**.

---

## 3. High-Level Features

1. **Tile Map**

   * 28 × 31 grid (standard Pac-Man maze dimensions).
   * Represented as a `.tmx` file (Tiled map editor) or ASCII array.
   * Tiles: wall, empty, pellet, power pellet, tunnel.
   * Wraparound tunnels connecting left/right edges.

2. **Entities**

   * **Player** (Pac-Man-like character):

     * Moves in four directions on grid.
     * Turns only when aligned to cell center.
   * **Ghosts** (4 types):

     * Each with unique target behavior.
     * States: Scatter, Chase, Frightened, Eaten.
   * **Fruits**:

     * Appears at specific times; collectible for points.

3. **Core Systems**

   * **Input handling** (keyboard arrows/WASD).
   * **Grid-based movement system** with queued direction.
   * **Collision detection** (pellet eating, ghost contact, wall blocking).
   * **Pathfinding** for ghosts (BFS or A\* on grid).
   * **State machine** for game flow:

     * READY → PLAY → LIFE\_LOST → GAME\_OVER → LEVEL\_COMPLETE.
   * **Scoring & lives tracking**.
   * **Level progression** (speed increases, reset pellets).

4. **UI**

   * HUD showing:

     * Score
     * High score (with player name)
     * Lives remaining
     * Level icons
   * Center "READY!" text at start.

5. **Art & Sound**

   * All assets must be CC0 or custom-made.
   * Sprites for player, ghosts, pellets, fruits, walls.
    * Sound effects for pellet eat, power-up, ghost eaten, death.
    * Audio disabled by default; enable with `PACMAN_ENABLE_AUDIO=1`.

---

## 4. Technical Requirements

* **Language:** Go 1.22+
* **Library:** `github.com/hajimehoshi/ebiten/v2`
* **Asset Loading:** PNG for sprites, WAV/MP3 for audio.
* **Fixed timestep:** 60 updates per second.
* **Code Structure:**

  ```
  /assets
    /images
    /sounds
  /pkg
    /game         # Game loop, state machine
    /entities     # Player, Ghost structs
    /map          # Tile map loader & renderer
    /pathfinding  # BFS/A* utilities
    /ui           # HUD drawing
  main.go         # Entry point
  ```

---

## 5. Milestones

### Milestone 1 — Basic Map & Player Movement — DONE

* Load map from 2D array or `.tmx` file.
* Draw walls, pellets, empty space.
* Move player with arrow keys; stop at walls.

### Milestone 2 — Pellets & Scoring — DONE

* Detect pellet collision.
* Increase score; remove pellet from map.

### Milestone 3 — Ghosts (Random Movement)

* Spawn ghosts in ghost house.
* Move randomly; collision with player ends life.

### Milestone 4 — Ghost AI & States

* Implement pathfinding to chase player.
* Add Scatter/Chase/Frightened modes with timers.

### Milestone 5 — Power Pellets & Frightened Mode

* Player eats power pellet → ghosts turn blue and run away.
* Eating ghosts sends them back to ghost house.

### Milestone 6 — Fruits & Level Progression

* Spawn fruit at specific times.
* New levels reset map and increase speed.

---

## 6. Stretch Features (Optional)

* Two-player alternating mode.
* Configurable maps.
* High-score persistence to file.

---

## 7. Non-Functional Requirements

* **Performance:** Stable 60 FPS.
* **Portability:** Runs on macOS, Windows, Linux.
* **Maintainability:** Modular design, small files, clear interfaces.
* **Readability:** Commented code, consistent naming.

---

## 8. Implementation Status

### Completed Features
- **Map & Rendering**: Implemented 28×31 ASCII maze with walls, pellets, power pellets, and wrap tunnels. Drawn at native resolution and scaled to fit ~75% of the display on first launch. Fullscreen toggle supported.
- **Player Movement**: 
  - Grid-based movement with queued direction and cell-center turns
  - Wall collision enforced with wrap-around horizontally
  - Current speed: 720 px/s at 60 UPS (1.5× original speed)
  - **Fixed**: Improved turn detection with dynamic alignment threshold for responsive controls at high speeds
- **Ghost System** (Milestone 3 ✓): 
  - 4 ghosts with random movement behavior
  - Ghosts spawn in ghost house area  
  - Ghost-player collision detection with life loss
  - Current speed: 630 px/s at 60 UPS (1.5× original speed)
- **Power Pellets & Frightened Mode** (Milestone 5 ✓):
  - Power pellets trigger frightened mode for 120 ticks (2 seconds at 60 UPS)
  - Ghosts turn blue and can be eaten during frightened mode
  - Score combo system: 200/400/800/1600 points for consecutive ghost eating
  - Eaten ghosts return to ghost house
  - Timer correctly expires and returns ghosts to normal state
  - Frightened timer display shows remaining seconds
- **Audio System** ✓:
  - Complete audio manager with support for pellet, power pellet, ghost eaten, and death sounds
  - Graceful fallback to synthesized beeps when sound files are missing
  - Disabled by default; enable with `PACMAN_ENABLE_AUDIO=1`
  - Sound files go in `assets/sounds/` (pellet.wav, power.wav, ghost.wav, death.wav)
- **High Score & Leaderboard** ✓:
  - Player name entry at game start (max 12 characters)
  - Persistent high score storage in JSON format
  - Multi-player leaderboard tracking all players' best scores
  - Leaderboard UI accessible via 'S' key or shown on game over
  - Legacy high score file import support
  - Atomic file writes for safe concurrent access
- **Input Controls**: 
  - Arrow keys to move
  - `Space` to pause
  - `F` to toggle fullscreen
  - `Q` to quit (shows leaderboard first)
  - `S` to toggle leaderboard
  - `R`/`Y` for easter eggs
- **Scoring**: Pellets +10, Power Pellets +50, Frightened ghosts 200-1600 (combo).
- **HUD**: 
  - Player name, Score, High score (with owner's name), Lives, and FPS counter in top-left
  - Frightened mode countdown timer in bottom-right (cyan, shows decimal seconds)
- **Easter Eggs**:
  - Name-based: Enter "Rekha" or "Roy" as name for special message
  - Key-based: Press 'R' or 'Y' during gameplay
  - Random: ~1/6000 chance per update (~100 second average) for surprise message
  - Messages appear in pink for 3 seconds
- **Project Structure**: 
  - `cmd/pacman`: Entry point
  - `internal/game`: Core game loop, audio manager, high score system
  - `internal/entities`: Player and ghost structs
  - `internal/tilemap`: Maze rendering and tile management
  - `internal/ui`: HUD utilities
  - `assets/sounds/`: Audio files (currently empty)
  - `CLAUDE.md`: AI development assistant instructions with detailed architecture
- **Build tooling**: 
  - Makefile with `deps`, `build`, `run`, `test`, `coverage`, `coverage-html`, `release` targets
  - Cross-platform builds for Linux, macOS, Windows
  - Comprehensive unit tests including audio and high score systems

### Technical Details
- **Game Speed**: 60 updates per second (UPS)
- **Player Speed**: 720 pixels/second (12 pixels per update)
- **Ghost Speed**: 630 pixels/second (10.5 pixels per update)  
- **Frightened Duration**: 120 ticks (2 seconds at 60 UPS)
- **Lives**: 3 lives, position reset on death
- **Alignment Threshold**: 6 pixels (`playerSpeedPixelsPerUpdate/2`) for responsive turning
- **High Score Storage**: `$HOME/.config/pacman/highscore.json` (or `PACMAN_CONFIG_DIR`)

### Known Issues
- **Current Issue**: Up/down arrow keys experiencing responsiveness issues (partial fixes attempted)

### Bug Fixes
- **Fixed**: Movement keys not responding when game paused or showing leaderboard
- **Fixed**: Direction changes not registering until hitting wall (improved alignment detection)
- **Fixed**: Frightened mode timer expiring correctly after timeout

### Next Up
- **Priority**: Fix up/down arrow key responsiveness issues
- Ghost AI with pathfinding (Milestone 4) - Implement chase/scatter modes
- Fruits and level progression (Milestone 6)
- Actual sound file assets