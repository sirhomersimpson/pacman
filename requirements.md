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
     * High score
     * Lives remaining
     * Level icons
   * Center "READY!" text at start.

5. **Art & Sound**

   * All assets must be CC0 or custom-made.
   * Sprites for player, ghosts, pellets, fruits, walls.
   * Sound effects for pellet eat, power-up, ghost eaten, death.

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

### Milestone 1 — Basic Map & Player Movement

* Load map from 2D array or `.tmx` file.
* Draw walls, pellets, empty space.
* Move player with arrow keys; stop at walls.

### Milestone 2 — Pellets & Scoring

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

- **Map & Rendering**: Implemented 28×31 ASCII maze with walls, pellets, power pellets, and wrap tunnels. Drawn at native resolution and scaled to fit ~75% of the display on first launch. Fullscreen toggle supported.
- **Player Movement**: Grid-based movement with queued direction and cell-center turns. Wall collision enforced. Wrap-around horizontally. Current speed: fast (480 px/s at 60 UPS).
- **Input**: Arrow keys to move, `Space` to pause, `F` to toggle fullscreen, `Q` to quit.
- **Scoring**: Pellets +10, Power Pellets +50. HUD shows Score top-left.
- **Project Structure**: `main.go`, `pkg/game`, `pkg/entities`, `pkg/tilemap`, `pkg/ui`, `assets/` directories created.
- **Build tooling**: Makefile with `deps`, `build`, `run`, `test`, `release` targets. Unit tests added for entities, tilemap, and basic game constraints.

### Next Up
- Ghosts (Milestone 3), Ghost AI & states (Milestone 4)
- Power mode behavior (Milestone 5 reactions) and fruits (Milestone 6)
- High score persistence and richer HUD