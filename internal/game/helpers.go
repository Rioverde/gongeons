package game

// This file defines the Stats struct and related functions for calculating derived stats based on core stats.
const (
	BaseHealthPerStrength   = 10
	BaseHealthPerDexterity  = 5
	BaseManaPerIntelligence = 10
)

func calculateHealth(strength, dexterity int) int {
	return strength*BaseHealthPerStrength + dexterity*BaseHealthPerDexterity
}

func calculateMana(intelligence int) int {
	return intelligence * BaseManaPerIntelligence
}
