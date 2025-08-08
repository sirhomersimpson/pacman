package tilemap

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Tile int

const (
	TileEmpty Tile = iota
	TileWall
	TilePellet
	TilePower
)

type TileMap struct {
	Width    int
	Height   int
	TileSize int
	Tiles    [][]Tile
}

func NewDefaultMap(tileSize int) *TileMap {
	grid := parseMaze(defaultMaze)
	return &TileMap{
		Width:    len(grid[0]),
		Height:   len(grid),
		TileSize: tileSize,
		Tiles:    grid,
	}
}

func (m *TileMap) IsWall(x, y int) bool {
	if y < 0 || y >= m.Height || x < 0 || x >= m.Width {
		return true
	}
	return m.Tiles[y][x] == TileWall
}

// EatPelletAt removes a pellet/power pellet at grid cell and returns (ate, power)
func (m *TileMap) EatPelletAt(x, y int) (bool, bool) {
	if y < 0 || y >= m.Height || x < 0 || x >= m.Width {
		return false, false
	}
	if m.Tiles[y][x] == TilePellet {
		m.Tiles[y][x] = TileEmpty
		return true, false
	}
	if m.Tiles[y][x] == TilePower {
		m.Tiles[y][x] = TileEmpty
		return true, true
	}
	return false, false
}

func (m *TileMap) Draw(dst *ebiten.Image) {
	blue := color.RGBA{R: 33, G: 33, B: 255, A: 255}
	pelletColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			t := m.Tiles[y][x]
			px := float32(x * m.TileSize)
			py := float32(y * m.TileSize)
			cx := px + float32(m.TileSize/2)
			cy := py + float32(m.TileSize/2)

			switch t {
			case TileWall:
				// Draw a filled rectangle for the wall
				rect := ebiten.NewImage(m.TileSize, m.TileSize)
				rect.Fill(blue)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(px), float64(py))
				dst.DrawImage(rect, op)
			case TilePellet:
				vector.DrawFilledCircle(dst, cx, cy, float32(m.TileSize)/8, pelletColor, true)
			case TilePower:
				vector.DrawFilledCircle(dst, cx, cy, float32(m.TileSize)/4, pelletColor, true)
			}
		}
	}
}

func parseMaze(lines []string) [][]Tile {
	h := len(lines)
	w := len(lines[0])
	grid := make([][]Tile, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			switch lines[y][x] {
			case '#':
				grid[y][x] = TileWall
			case '.':
				grid[y][x] = TilePellet
			case 'o':
				grid[y][x] = TilePower
			default:
				grid[y][x] = TileEmpty
			}
		}
	}
	return grid
}
