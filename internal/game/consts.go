package game

type Slot string

const (
	// Game constants
	head Slot = "head"
	body Slot = "body"
	legs Slot = "legs"

	// Damage multipliers for different body parts
	Head                 = 2.0
	bodyDamageMultiplier = 1.0
	legsDamageMultiplier = 0.5
	numberOfSlots        = 3
)
