package game

import (
	"github.com/legendary-code/hexe/pkg/hexe"
	"github.com/legendary-code/hexe/pkg/hexe/coord"
)

// World is a wrapper around hexe.AxialGrid that exposes a game-level API
// (plain ints as coordinates) and keeps hexe types from leaking to callers.
type World struct {
	grid hexe.AxialGrid[Tile]
}

// NewWorld creates an empty game world backed by an axial hex grid.
func NewWorld() *World {
	return &World{
		grid: hexe.NewAxialGrid[Tile](),
	}
}

// GetTile returns the tile at (q, r) and whether it exists.
// Uses hexe's Index method (Get returns only the value, no existence flag).
func (w *World) GetTile(q, r int) (Tile, bool) {
	return w.grid.Index(coord.NewAxial(q, r))
}

// SetTile places a tile at (q, r), replacing any existing tile there.
func (w *World) SetTile(q, r int, tile Tile) {
	w.grid.Set(coord.NewAxial(q, r), tile)
}
