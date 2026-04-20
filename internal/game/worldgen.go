package game

import (
	"math"

	"github.com/legendary-code/hexe/pkg/hexe/coord"
	opensimplex "github.com/ojrac/opensimplex-go"
)

// BiomeRule maps an upper elevation bound to a terrain type.
// Rules are applied in order from lowest to highest MaxElevation:
// a tile is assigned the first Terrain whose MaxElevation strictly exceeds the tile's elevation.
type BiomeRule struct {
	MaxElevation float64
	Terrain      Terrain
}

// DefaultBiomes is an elevation-ordered ruleset from water up to mountains.
// The last rule uses MaxElevation > 1.0 as a catch-all for the highest peaks.
var DefaultBiomes = []BiomeRule{
	{MaxElevation: 0.30, Terrain: TerrainWater},
	{MaxElevation: 0.35, Terrain: TerrainJungle},
	{MaxElevation: 0.45, Terrain: TerrainDirt},
	{MaxElevation: 0.55, Terrain: TerrainMeadow},
	{MaxElevation: 0.65, Terrain: TerrainGrass},
	{MaxElevation: 0.80, Terrain: TerrainForest},
	{MaxElevation: 1.01, Terrain: TerrainMountain},
}

// GenerateConfig controls procedural world generation.
type GenerateConfig struct {
	// Radius is the hex radius from the center. The resulting map is a large hexagon of this size.
	Radius int

	// Seed makes generation deterministic — the same seed always produces the same map.
	Seed int64

	// Scale is the noise coordinate scale. Larger values stretch the noise out and create bigger
	// continents; smaller values break the land into more islets. A reasonable starting range is 8..15.
	Scale float64

	// Falloff is the edge attenuation exponent that creates an island shape. Zero disables the
	// attenuation entirely and lets the noise reach the border. Values around 2..3 yield soft coastlines.
	Falloff float64

	// Biomes maps elevation ranges to terrain types. When empty, DefaultBiomes is used.
	Biomes []BiomeRule
}

// GenerateWorld fills w with procedurally generated tiles over a hex region of the given radius.
// For each coordinate it samples a Simplex noise value in [0, 1], and when Falloff is positive
// multiplies that value by (1 - dist/radius)^Falloff so the edges of the map sink into water and
// the terrain takes an island shape. The resulting elevation is then passed through the biome
// rules and the matching tile is placed at that coordinate.
func GenerateWorld(w *World, cfg GenerateConfig) {
	if cfg.Scale == 0 {
		cfg.Scale = 12.0
	}
	if len(cfg.Biomes) == 0 {
		cfg.Biomes = DefaultBiomes
	}

	noise := opensimplex.NewNormalized(cfg.Seed)
	center := coord.ZeroAxial()
	radius := float64(cfg.Radius)

	coords := center.MovementRange(cfg.Radius)
	for it := coords.Iterator(); it.Next(); {
		c := it.Item()
		q, r := c.Q(), c.R()

		elevation := noise.Eval2(float64(q)/cfg.Scale, float64(r)/cfg.Scale)

		if cfg.Falloff > 0 && cfg.Radius > 0 {
			dist := float64(c.DistanceTo(center))
			t := min(dist/radius, 1.0)
			elevation *= math.Pow(1.0-t, cfg.Falloff)
		}

		w.SetTile(q, r, Tile{Terrain: pickTerrain(elevation, cfg.Biomes)})
	}
}

// Generate is a convenience wrapper with sensible defaults: an island with soft edge falloff.
func (w *World) Generate(radius int, seed int64) {
	GenerateWorld(w, GenerateConfig{
		Radius:  radius,
		Seed:    seed,
		Scale:   12.0,
		Falloff: 2.0,
	})
}

// pickTerrain returns the first Terrain whose MaxElevation threshold strictly exceeds elevation.
// Rules must be sorted by ascending MaxElevation (which is true for DefaultBiomes).
func pickTerrain(elevation float64, rules []BiomeRule) Terrain {
	for _, rule := range rules {
		if elevation < rule.MaxElevation {
			return rule.Terrain
		}
	}
	return rules[len(rules)-1].Terrain
}
