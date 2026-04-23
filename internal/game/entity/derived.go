package entity

// DerivedStats is the combat-tick state shared by every playable entity.
// It holds computed pools (HP, Mana) and scheduling state (Speed, Energy,
// Initiative, Intent) that are seeded from CoreStats at construction time
// and mutated in place as the simulation ticks. Any type that embeds
// DerivedStats automatically picks up TakeDamage, IsAlive, and Equip;
// Player and Monster use this to avoid duplicating the same fields and
// methods on both sides of the combat loop.
type DerivedStats struct {
	Equipment map[Slot]*Armor `json:"equipment,omitempty"`

	MaxHP   int `json:"max_hp"`
	HP      int `json:"hp"`
	MaxMana int `json:"max_mana"`
	Mana    int `json:"mana"`

	Speed      int `json:"speed"`
	Energy     int `json:"energy"`
	Initiative int `json:"initiative"`
	Intent     any `json:"-"`
}

// TakeDamage applies incoming damage, passing it through each equipped
// armor piece in turn. A dead entity (HP <= 0) is a no-op; damage that
// survives all absorbers is subtracted from HP and clamped at zero.
// Entities with a nil Equipment map (the common case for raw monsters)
// take the full damage directly.
func (d *DerivedStats) TakeDamage(damage int) {
	if !d.IsAlive() {
		return
	}
	for _, armor := range d.Equipment {
		if damage <= 0 {
			break
		}
		if armor != nil {
			damage = armor.Reduce(damage)
		}
	}
	d.HP -= damage
	d.HP = max(0, d.HP)
}

// IsAlive reports whether the entity's HP is above zero.
func (d *DerivedStats) IsAlive() bool {
	return d.HP > 0
}

// Equip puts armor into the named slot, replacing whatever was there.
// Lazily initializes the Equipment map so callers can Equip on a
// freshly-constructed entity without touching the field directly.
func (d *DerivedStats) Equip(slot Slot, armor *Armor) {
	if d.Equipment == nil {
		d.Equipment = make(map[Slot]*Armor, NumberOfSlots)
	}
	d.Equipment[slot] = armor
}
