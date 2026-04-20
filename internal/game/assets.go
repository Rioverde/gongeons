package game

// TileAsset returns the filename of the PNG tile that represents the given terrain.
// Filenames are relative to the tiles asset directory (e.g. "assets/tiles/water.png").
// An unknown terrain falls back to dirt so the caller never gets an empty string.
func TileAsset(t Terrain) string {
	switch t {
	case TerrainWater:
		return "water.png"
	case TerrainJungle:
		return "jungle.png"
	case TerrainDirt:
		return "dirt.png"
	case TerrainMeadow:
		return "meadow.png"
	case TerrainGrass:
		return "grass.png"
	case TerrainForest:
		return "forest.png"
	case TerrainMountain:
		return "mountain.png"
	case TerrainCursedForest:
		return "forest.png"
	default:
		return "dirt.png"
	}
}
