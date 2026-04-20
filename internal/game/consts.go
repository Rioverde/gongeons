package game

type Slot string
type Terrain string

const (
	// Game constants
	SlotHead Slot = "head"
	SlotBody Slot = "body"
	SlotLegs Slot = "legs"

	TerrainGrass        Terrain = "grass"
	TerrainDirt         Terrain = "dirt"
	TerrainWater        Terrain = "water"
	TerrainMountain     Terrain = "mountain"
	TerrainForest       Terrain = "forest"
	TerrainJungle       Terrain = "jungle"
	TerrainCursedForest Terrain = "cursed_forest"
	TerrainMeadow       Terrain = "meadow"

	// Damage multipliers for different body parts
	HeadDamageMultiplier = 2.0
	BodyDamageMultiplier = 1.0
	LegsDamageMultiplier = 0.5
	numberOfSlots        = 3
)
