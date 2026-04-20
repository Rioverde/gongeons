package game

type Tile struct {
	Terrain  Terrain  `json:"terrain"`
	Occupant Occupant `json:"occupant,omitempty"`
}
