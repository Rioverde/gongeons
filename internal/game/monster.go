package game

import "errors"

// Ensure that Monster implements the Combatant interface at compile time.
var _ Combatant = (*Monster)(nil)

// Monster represents a monster in the game with an ID, name, and stats.
type Monster struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Stats *Stats `json:"stats"`
}

// TakeDamage applies damage to the monster's health, ensuring that health does not drop below zero.
func (m *Monster) TakeDamage(damage int) {
	m.Stats.applyDamage(damage)
}

// NewMonster creates a new monster with the given parameters and returns a pointer to the Monster struct.
func NewMonster(id, name string, strength, dexterity, intelligence int) (*Monster, error) {
	// In a real application, you might want to add validation for the input parameters here.
	if id == "" {
		return nil, errors.New("invalid monster ID")
	}
	// For simplicity, we just check if the name is empty. You can add more complex validation as needed.
	if name == "" {
		return nil, errors.New("invalid monster name")
	}
	// Ensure that core stats are not negative. You can adjust this logic based on your game's requirements.
	if strength < 0 || dexterity < 0 || intelligence < 0 {
		return nil, errors.New("core stats cannot be negative")
	}
	// Create a new Stats struct for the monster using the provided core stats and return the complete Monster struct.
	return &Monster{
		ID:    id,
		Name:  name,
		Stats: NewStats(strength, dexterity, intelligence),
	}, nil
}

// IsAlive checks if the monster is still alive (i.e., has health greater than zero).
func (m *Monster) IsAlive() bool {
	return m.Stats.Health > 0
}

// GetStats returns the current stats of the monster, including health and mana.
func (m *Monster) GetStats() Stats {
	return *m.Stats
}
