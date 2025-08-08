# Pac-Man (Go + Ebiten)

## Overview

A feature-rich Pac-Man clone written in Go using the Ebiten game engine. This implementation includes:
- Classic 28×31 ASCII-based maze with authentic gameplay
- Smooth grid-based movement with improved turn detection  
- Four ghosts with random movement patterns
- Power pellets with 2-second frightened mode and scoring combos
- Persistent high scores with multi-player leaderboard
- Audio system with synthesized fallback sounds
- Hidden easter eggs and personal touches

## Quick Start

### Prerequisites
- Go 1.19+ (1.22+ recommended)
- Make (optional but recommended)

### Build & Run
```bash
# macOS/Linux
make run

# Windows
make build
./bin/pacman.exe

# Without make
go build -o pacman ./cmd/pacman
./pacman
```

## Controls

| Key | Action |
|-----|--------|
| **Arrow Keys** | Move Pacman |
| **Space** | Pause/Resume |
| **F** | Toggle fullscreen |
| **S** | Show/Hide leaderboard |
| **Q** | Quit (shows leaderboard first) |
| **R** | Easter egg: "Dad Loves Rekha" |
| **Y** | Easter egg: "Dad Loves Roy" |

## Gameplay Features

### Scoring System
- **Regular pellets**: 10 points each
- **Power pellets**: 50 points each
- **Frightened ghosts**: 200 → 400 → 800 → 1600 points (combo multiplier)

### Power Mode
- Power pellets activate frightened mode for exactly 2 seconds (120 ticks)
- Ghosts turn blue and can be eaten for bonus points
- Score multiplier increases with each ghost eaten in sequence
- Eaten ghosts return to the ghost house

### Name Entry & High Scores
- Enter your name at game start (max 12 characters: letters, numbers, spaces, _, -)
- High scores are saved per player in JSON format
- Storage location: `$HOME/.config/pacman/highscore.json`
- Leaderboard shows top 10 players, accessible via 'S' key or on game over
- Legacy high score file import supported

### Audio System
Audio is **disabled by default**. To enable:

```bash
# Enable for single run
PACMAN_ENABLE_AUDIO=1 make run

# Force disable (overrides enable)
PACMAN_DISABLE_AUDIO=1 make run
```

**Sound Files**: Place WAV files in `assets/sounds/`:
- `pellet.wav` - Pellet collection sound
- `power.wav` - Power pellet sound  
- `ghost.wav` - Ghost eaten sound
- `death.wav` - Player death sound

If files are missing, the game synthesizes simple beep sounds as fallbacks.

### Easter Eggs
- **Name-based**: Enter "Rekha" or "Roy" as your name for a special message
- **Key-based**: Press 'R' or 'Y' during gameplay for instant messages
- **Random**: ~1% chance every 100 seconds for surprise messages
- All messages appear in pink for 3 seconds

## Development

### Testing
```bash
make test           # Run all tests
make coverage       # Generate coverage report
make coverage-html  # Generate HTML coverage report
```

### Code Quality
```bash
make fmt    # Format code
make vet    # Run go vet
make deps   # Update dependencies
```

### Cross-Platform Builds
```bash
make release        # Build for all platforms
make build-linux    # Linux build
make build-darwin   # macOS build  
make build-windows  # Windows build
```

## Project Structure

```
├── cmd/pacman/          # Entry point
├── internal/
│   ├── game/           # Core game logic, audio, high scores
│   ├── entities/       # Player and ghost definitions
│   ├── tilemap/        # Maze rendering and tile management
│   └── ui/             # HUD utilities
├── assets/
│   └── sounds/         # Audio files (currently empty)
├── Makefile           # Build automation
└── requirements.md    # Detailed implementation status
```

## Technical Details

- **Game Speed**: 60 updates per second (UPS)
- **Player Speed**: 720 pixels/second (1.5× original speed)
- **Ghost Speed**: 630 pixels/second (1.5× original speed)
- **Movement**: Grid-based with 6-pixel alignment threshold for responsive turning
- **Resolution**: Native 28×31 tile maze, auto-scaled to fit ~75% of display
- **Persistence**: High scores stored in OS user config directory

## Recent Bug Fixes

- ✅ **Fixed**: Movement keys not responding when game paused or showing leaderboard
- ✅ **Fixed**: Direction changes not registering until hitting wall (improved alignment detection)
- ✅ **Fixed**: Frightened mode timer expiring correctly after timeout

## License

See `LICENSE` file for details.

---

*A tribute project with personal touches and professional polish.*