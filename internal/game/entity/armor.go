package entity

// Armor is a single equippable armor piece providing flat per-slot damage
// reduction. Defense is subtracted from incoming damage by Armor.Reduce;
// negative results clamp to zero. Armor pieces are stored in a
// Character's DerivedStats.Equipment map, keyed by Slot.
type Armor struct {
	Name        string `json:"name"`
	Defense     int    `json:"defense"`
	Description string `json:"description,omitempty"`
}

// Reduce returns the damage left after this armor piece absorbs its
// share. Never returns a negative value.
func (a *Armor) Reduce(damage int) int {
	return max(0, damage-a.Defense)
}
