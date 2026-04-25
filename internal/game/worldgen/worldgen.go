// Package worldgen generates the bounded, mapgen2-style game world.
//
// The real generation pipeline lives in world.go's Generate function;
// per-feature sources (regions, landmarks, volcanoes, deposits) wrap the
// produced *World in dedicated files. This file holds the InfluenceSampler
// interface plus its zero-state implementation, which the client UI
// consumes for cosmetic mini-map tinting before the server pushes real
// region samples.
package worldgen

import (
	"github.com/Rioverde/gongeons/internal/game/world"
)

// InfluenceSampler is the per-tile thematic-influence lookup the client
// UI consumes for region tinting. The real RegionSource implements it,
// but the client side instantiates a zero-state sampler at join time so
// the UI can render before any region data arrives over the wire.
type InfluenceSampler interface {
	InfluenceAt(x, y int) world.RegionInfluence
}

// zeroInfluenceSampler reports zero influence for every coordinate. The
// client UI uses it as a placeholder until server-pushed region data
// flows in; mini-map tints render as neutral in the meantime.
type zeroInfluenceSampler struct{}

// NewInfluenceSampler returns a sampler that reports zero influence
// everywhere. Used by the client UI before the wire protocol carries
// region samples; once the server pushes real region data, the mini-map
// tint layer reads from that instead.
func NewInfluenceSampler(_ int64) InfluenceSampler { return zeroInfluenceSampler{} }

// InfluenceAt always returns the zero RegionInfluence (all fields 0).
func (zeroInfluenceSampler) InfluenceAt(_, _ int) world.RegionInfluence {
	return world.RegionInfluence{}
}

