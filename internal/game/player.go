package game

import (
	"errors"
)

// Ensure that Player implements the Combatant interface at compile time.
var _ Combatant = (*Player)(nil)

// Player represents a player in the game with an ID, name, and stats.
type Player struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Stats     *Stats          `json:"stats"`
	Equipment map[Slot]*Armor `json:"equipment,omitempty"`
}

// Armor represents the armor equipped by the player, which can provide additional defense.
type Armor struct {
	Name        string `json:"name"`
	Defense     int    `json:"defense"`
	Description string `json:"description,omitempty"`
}

// TakeDamage applies damage to the player's health, considering the defense provided by equipped armor pieces.
func (p *Player) TakeDamage(damage int) {
	// If the player is already at or below zero health, stop processing damage.
	if !p.IsAlive() {
		return
	}
	// Calculate the total defense provided by all equipped armor pieces.
	for _, armor := range p.Equipment {
		if damage <= 0 {
			break // No more damage to apply, exit the loop.
		}
		if armor != nil {
			// Pass the damage through each armor piece; Reduce returns the damage left after this piece absorbs its share.
			damage = armor.Reduce(damage)
		}
	}
	// Apply the remaining damage to the player's health.
	p.Stats.applyDamage(damage)
}

// Reduce calculates the effective damage after applying the armor's defense and ensures that the damage does not go below zero.
func (a *Armor) Reduce(damage int) int {
	return max(0, damage-a.Defense)
}

func (p *Player) IsAlive() bool {
	// A player is considered alive if their health is greater than zero.
	return p.Stats.Health > 0
}

func (p *Player) GetStats() Stats {
	// Return the current stats of the player, including health and mana.
	return *p.Stats
}

// Equip allows the player to equip an armor piece in a specified slot, replacing any existing armor in that slot.
func (p *Player) Equip(slot Slot, armor *Armor) {
	p.Equipment[slot] = armor
}

// NewPlayer creates a new player with the given parameters and returns a pointer to the Player struct.
func NewPlayer(id, name string, strength, dexterity, intelligence int) (*Player, error) {
	// In a real application, you might want to add validation for the input parameters here.
	if id == "" {
		return nil, errors.New("invalid player ID")
	}
	// For simplicity, we just check if the name is empty. You can add more complex validation as needed.
	if name == "" {
		return nil, errors.New("invalid player name")
	}
	// Ensure that core stats are not negative. You can adjust this logic based on your game's requirements.
	if strength < 0 || dexterity < 0 || intelligence < 0 {
		return nil, errors.New("core stats cannot be negative")
	}

	// Create and return the new player with the calculated stats.
	return &Player{
		ID:        id,
		Name:      name,
		Stats:     NewStats(strength, dexterity, intelligence),
		Equipment: make(map[Slot]*Armor, numberOfSlots),
	}, nil
}
