# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build & Run
```bash
make run                # Build and run the game
make build              # Build binary to bin/pacman
make clean              # Remove build artifacts
```

### Development
```bash
make test               # Run all unit tests
make coverage           # Generate coverage report
make coverage-html      # Generate HTML coverage report
make fmt                # Format code
make vet                # Run go vet
make deps               # Update dependencies (go mod tidy)
```

### Cross-Platform Builds
```bash
make release            # Build for all platforms
make build-linux        # Linux build
make build-darwin       # macOS build
make build-windows      # Windows build
```

### Running Tests
```bash
go test ./...                           # All tests
go test ./internal/game                 # Game package tests only
go test -run TestFrightenedModeTimeout  # Single test
go test -v ./internal/game              # Verbose output
```

## Architecture Overview

### Game Loop Structure
The game follows a classic game loop pattern:
- `Update()`: Game state updates (60 UPS) - handles input, movement, collision, timers
- `Draw()`: Rendering logic - draws to offscreen buffer then scales to display
- `Layout()`: Screen dimension management

### Core Systems

**Movement System**: Grid-based with queued turns. Player movement uses alignment threshold (`playerSpeedPixelsPerUpdate/2`) to enable responsive turning at high speeds. The `canMove()` function enforces alignment for movement, while `canMoveFromCellCenter()` allows turn queueing.

**Timer System**: Uses `tickCounter` for all timing. Key timers:
- `frightenedUntilTick`: Controls power pellet mode duration (120 ticks = 2 seconds)
- `easterUntilTick`: Controls easter egg message display (3 seconds)

**Audio System**: Lazy-loaded `AudioManager` with graceful fallback to synthesized beeps. Disabled by default - enable with `PACMAN_ENABLE_AUDIO=1`. Files expected in `assets/sounds/`.

**High Score System**: JSON persistence to user config directory with atomic writes. Supports multiple players and legacy text file import. Always save high scores immediately when achieved, not just on quit.

### State Management
Game has several states that affect input handling:
- `enteringName`: Name entry at startup
- `showingLeaderboard`: Leaderboard display
- `paused`: Game paused
- Movement input is blocked when not in active play state

### Key Implementation Details

**Frightened Mode**: 
- Duration: exactly 120 ticks (not 6 seconds - use tick-based timing)
- Timer check runs before game logic in `Update()`
- Combo scoring: 200 → 400 → 800 → 1600 points, resets on new power pellet

**Grid Alignment**: 
- Tiles are 16x16 pixels
- Cell centers are at `(x*16 + 8, y*16 + 8)`
- Alignment threshold is `playerSpeedPixelsPerUpdate/2` (6 pixels) for responsive controls
- When blocked, players snap to cell center to prevent jitter

**Easter Eggs**: 
- Name-based triggers on "rekha"/"roy" (case-insensitive)
- Random triggers ~1/6000 chance per update (~100 second average)
- Key-based triggers on 'R'/'Y' keys
- All display for 3 seconds in pink color

### Speed Configuration
Current speeds (1.5× original):
- Player: 720 px/s (12 px per update)
- Ghosts: 630 px/s (10.5 px per update)
- These speeds require the increased alignment threshold for responsive turning

### Environment Variables
- `PACMAN_ENABLE_AUDIO=1`: Enable audio system
- `PACMAN_DISABLE_AUDIO=1`: Force disable audio (takes precedence)
- `PACMAN_CONFIG_DIR`: Override high score storage location (for testing)

### Testing Notes
- Use `t.Setenv("PACMAN_CONFIG_DIR", t.TempDir())` for isolated high score tests
- Use `t.Setenv("PACMAN_DISABLE_AUDIO", "1")` to avoid audio context conflicts in tests
- Test alignment with values beyond threshold (8+ pixels offset)
- Movement system tests should account for the 6-pixel alignment threshold