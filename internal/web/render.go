package web

import (
	"math"

	"github.com/Rioverde/gongeons/internal/game"
)

// Tile PNG dimensions in pixels. The hex base fills the full image width and about 296px of the
// height; the remaining vertical space is headroom for objects that may spill above the hex.
const (
	tileImageWidth  = 256
	tileImageHeight = 384
	hexBaseHeight   = 296
)

// rowSpacing is the vertical distance between hex centers of adjacent rows. For a pointy-top hex
// of width W the full height is W * 2/sqrt(3); rows overlap by 1/4 so spacing is W/sqrt(3) * 1.5.
var rowSpacing = int(math.Round(float64(tileImageWidth) / math.Sqrt(3) * 1.5))

// axialToPixel converts axial hex coordinates into the top-left pixel position of a PNG tile,
// assuming pointy-top layout and the tile dimensions declared above.
func axialToPixel(q, r int) (int, int) {
	left := tileImageWidth*q + tileImageWidth*r/2
	top := rowSpacing*r - (tileImageHeight - hexBaseHeight)
	return left, top
}

type tileView struct {
	Left  int
	Top   int
	File  string
	Depth int
}

type viewModel struct {
	Tiles    []tileView
	Width    int
	Height   int
	Seed     int64
	TileImgW int
	TileImgH int
}

// buildViewModel walks the world, normalizes tile positions so the map starts at (0, 0), and
// returns the data needed by the HTML template.
func buildViewModel(world *game.World, seed int64) viewModel {
	var (
		tiles                                  []tileView
		minLeft, minTop                        = math.MaxInt, math.MaxInt
		maxRight, maxBottom                    = math.MinInt, math.MinInt
	)

	world.ForEach(func(q, r int, t game.Tile) {
		left, top := axialToPixel(q, r)
		tiles = append(tiles, tileView{
			Left:  left,
			Top:   top,
			File:  game.TileAsset(t.Terrain),
			Depth: top,
		})
		minLeft = min(minLeft, left)
		minTop = min(minTop, top)
		maxRight = max(maxRight, left+tileImageWidth)
		maxBottom = max(maxBottom, top+tileImageHeight)
	})

	for i := range tiles {
		tiles[i].Left -= minLeft
		tiles[i].Top -= minTop
		tiles[i].Depth = tiles[i].Top
	}

	return viewModel{
		Tiles:    tiles,
		Width:    maxRight - minLeft,
		Height:   maxBottom - minTop,
		Seed:     seed,
		TileImgW: tileImageWidth,
		TileImgH: tileImageHeight,
	}
}
