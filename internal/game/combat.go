package game

type Combatant interface {
	// TakeDamage applies damage to the combatant, reducing their health accordingly.
	TakeDamage(damage int)
	// IsAlive checks if the combatant is still alive (i.e., has health greater than zero).
	IsAlive() bool
	// GetStats returns the current stats of the combatant, including health and mana.
	GetStats() Stats
}
