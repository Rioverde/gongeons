package worldgen

import (
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/Rioverde/gongeons/internal/game/geom"
	"github.com/Rioverde/gongeons/internal/game/naming"
	gworld "github.com/Rioverde/gongeons/internal/game/world"
)

// Per-super-chunk landmark count distribution. The four buckets sum to
// 1.0 and yield ~0.5 landmarks per super-chunk on average — sparse
// enough that players go many tiles between encounters yet dense enough
// that a Standard world still ships several hundred named landmarks.
const (
	landmarkProb0 float64 = 0.60 // 60% empty super-chunks
	landmarkProb1 float64 = 0.30 // 30% one landmark
	landmarkProb2 float64 = 0.08 // 8% two landmarks
	// Implicit landmarkProb3 = 0.02 — the remaining tail.

	// landmarkPlacementAttempts caps the per-landmark reject loop. With
	// minSpacing 4 and a 64-tile super-chunk a fresh draw lands on a
	// valid tile in 1-2 tries on average; 8 attempts is generous enough
	// that ocean / volcano / spacing rejections never starve the slot
	// silently. After exhausting the budget we drop the slot rather than
	// loop forever.
	landmarkPlacementAttempts = 8

	// landmarkMinSpacing is the minimum Chebyshev gap (in tiles) between
	// any two landmarks placed in the same super-chunk. Keeps clusters
	// from forming visible "landmark fields" that read as noise rather
	// than features.
	landmarkMinSpacing = 4
)

// landmarkSeedSalt mixes the world seed into a dedicated PRNG stream
// for landmark count + position rolls. Decouples landmark placement
// from regions, volcanoes, naming, and rivers so an unrelated tuning
// change to one of those subsystems does not shift landmark coords.
// Value: fractional hex of sqrt(41).
const landmarkSeedSalt int64 = 0x73c1a8f5b2e9d406

// landmarkPrefixCount mirrors the active.*.toml entries under
// "landmark.prefix.<character>.<idx>". All seven characters carry five
// prefixes today; if a future locale shrinks one bucket, the value
// here must drop in lock-step or naming.Generate will produce out-of-
// range PrefixIndex values.
var landmarkPrefixCount = map[string]int{
	"normal":   5,
	"blighted": 5,
	"fey":      5,
	"ancient":  5,
	"savage":   5,
	"holy":     5,
	"wild":     5,
}

// landmarkPatternCount mirrors the active.*.toml entries under
// "landmark.name.<sub_kind>.kind_pattern.<idx>". Counts per sub-kind:
// tower=3, giant_tree=2, standing_stones=2, obelisk=2, chasm=3, shrine=3.
var landmarkPatternCount = map[string]int{
	"landmark.tower":           3,
	"landmark.giant_tree":      2,
	"landmark.standing_stones": 2,
	"landmark.obelisk":         2,
	"landmark.chasm":           3,
	"landmark.shrine":          3,
}

// landmarkBounds returns a freshly built naming.Bounds. Maps are
// returned directly — naming.Generate only reads them, and the
// package-level values are never mutated after init, so sharing the
// same maps across every call is safe.
func landmarkBounds() naming.Bounds {
	return naming.Bounds{
		PatternCount: landmarkPatternCount,
		PrefixCount:  landmarkPrefixCount,
	}
}

// kindWeight pairs a LandmarkKind with its weight inside a per-character
// distribution. Weights are float32 cumulative probabilities — they are
// normalized at draw time, so the absolute values only need the
// relative ratio called out in the design doc.
type kindWeight struct {
	Kind   gworld.LandmarkKind
	Weight float32
}

// landmarkKindWeights resolves a region character to its weighted kind
// distribution. The mapping reads as a thematic affinity table:
// Blighted regions favour Obelisks and Chasms, Holy regions favour
// Shrines and Towers, Wild regions are full of Trees and Chasms, etc.
// Normal regions get a flat distribution across the four "civilised"
// kinds — Towers, Trees, Stones, Shrines.
//
// Tables are package-level fixed slices (not maps) because the lookup
// is hot, the set is small, and slice iteration is allocation-free.
var (
	landmarkKindsNormal = []kindWeight{
		{gworld.LandmarkTower, 0.25},
		{gworld.LandmarkGiantTree, 0.25},
		{gworld.LandmarkStandingStones, 0.25},
		{gworld.LandmarkShrine, 0.25},
	}
	landmarkKindsBlighted = []kindWeight{
		{gworld.LandmarkObelisk, 0.40},
		{gworld.LandmarkChasm, 0.30},
		{gworld.LandmarkStandingStones, 0.20},
		{gworld.LandmarkShrine, 0.10},
	}
	landmarkKindsFey = []kindWeight{
		{gworld.LandmarkGiantTree, 0.40},
		{gworld.LandmarkStandingStones, 0.30},
		{gworld.LandmarkShrine, 0.30},
	}
	landmarkKindsAncient = []kindWeight{
		{gworld.LandmarkObelisk, 0.50},
		{gworld.LandmarkStandingStones, 0.30},
		{gworld.LandmarkTower, 0.20},
	}
	landmarkKindsSavage = []kindWeight{
		{gworld.LandmarkChasm, 0.40},
		{gworld.LandmarkObelisk, 0.30},
		{gworld.LandmarkStandingStones, 0.30},
	}
	landmarkKindsHoly = []kindWeight{
		{gworld.LandmarkShrine, 0.50},
		{gworld.LandmarkTower, 0.30},
		{gworld.LandmarkGiantTree, 0.20},
	}
	landmarkKindsWild = []kindWeight{
		{gworld.LandmarkGiantTree, 0.40},
		{gworld.LandmarkChasm, 0.30},
		{gworld.LandmarkStandingStones, 0.30},
	}
)

// kindWeightsFor returns the weighted distribution for the given
// region character. Unknown characters fall back to the Normal table —
// defensive only; every character defined in the world package has an
// explicit entry above.
func kindWeightsFor(c gworld.RegionCharacter) []kindWeight {
	switch c {
	case gworld.RegionNormal:
		return landmarkKindsNormal
	case gworld.RegionBlighted:
		return landmarkKindsBlighted
	case gworld.RegionFey:
		return landmarkKindsFey
	case gworld.RegionAncient:
		return landmarkKindsAncient
	case gworld.RegionSavage:
		return landmarkKindsSavage
	case gworld.RegionHoly:
		return landmarkKindsHoly
	case gworld.RegionWild:
		return landmarkKindsWild
	}
	return landmarkKindsNormal
}

// LandmarkSource is the production world.LandmarkSource implementation.
// Per-super-chunk landmark sets are computed lazily on first
// LandmarksIn call and cached in a sync.Map. Each super-chunk consults
// its region (for character → kind weights) and optionally a
// VolcanoSource (to skip tiles inside core/slope rings).
//
// All fields are read-only after construction, with computed entries
// stored exclusively through sync.Map's atomic LoadOrStore. Concurrent
// LandmarksIn callers may both compute the same super-chunk before the
// cache fills — the duplicated work is bounded to the cache miss
// window and is harmless because the computation is pure.
type LandmarkSource struct {
	world     *World
	seed      int64
	regions   gworld.RegionSource
	volcanoes gworld.VolcanoSource // optional — may be nil

	// volcanoZoneAt is a flat lookup of every tile inside any volcano's
	// core or slope ring. Built once at construction so the per-tile
	// reject check is a single map read instead of a slice scan over
	// every nearby volcano. Nil when volcanoes is nil.
	volcanoZoneAt map[uint64]struct{}

	cache sync.Map // SuperChunkCoord -> []gworld.Landmark
}

// NewLandmarkSource builds a real landmark source over a finished
// worldgen.World. Replaces the all-empty NoiseLandmarkSource stub.
//
// volcanoes is optional — pass nil to disable the volcano-zone reject
// check. When non-nil, every tile inside a volcano's CoreTiles or
// SlopeTiles becomes ineligible for landmark placement. AshlandTiles
// stay eligible: ashland is dusted but passable terrain and reads
// thematically as a place where ancient things might still stand.
func NewLandmarkSource(
	w *World,
	seed int64,
	regions gworld.RegionSource,
	volcanoes gworld.VolcanoSource,
) *LandmarkSource {
	src := &LandmarkSource{
		world:     w,
		seed:      seed,
		regions:   regions,
		volcanoes: volcanoes,
	}
	if volcanoes != nil {
		src.volcanoZoneAt = buildVolcanoZoneIndex(volcanoes)
	}
	return src
}

// buildVolcanoZoneIndex walks every volcano via the source's per-super-
// chunk lookup and folds core+slope tiles into a flat reject set. Built
// once at construction so the per-placement reject check is a single
// map read. Ashland tiles are intentionally excluded — see
// NewLandmarkSource.
func buildVolcanoZoneIndex(volcanoes gworld.VolcanoSource) map[uint64]struct{} {
	out := map[uint64]struct{}{}
	// Walk every super-chunk a volcano could anchor in. We don't know the
	// exact set up front; sentinel-source signature is per-sc lookup so
	// we collect by re-querying with each unique anchor we hit. The
	// concrete VolcanoSource implementation memoises its volcanoes
	// slice, so this is safe and cheap on the standard implementation.
	if all, ok := volcanoes.(interface {
		All() []gworld.Volcano
	}); ok {
		for _, v := range all.All() {
			for _, t := range v.CoreTiles {
				out[packPos(t)] = struct{}{}
			}
			for _, t := range v.SlopeTiles {
				out[packPos(t)] = struct{}{}
			}
		}
	}
	return out
}

// LandmarksIn returns the deterministic landmark slice for sc.
// Concurrent callers may both compute on a cache miss; sync.Map's
// LoadOrStore makes the second writer's result visible atomically and
// the duplicated work is bounded to the miss window.
func (s *LandmarkSource) LandmarksIn(sc geom.SuperChunkCoord) []gworld.Landmark {
	if cached, ok := s.cache.Load(sc); ok {
		return cached.([]gworld.Landmark)
	}
	out := s.computeLandmarks(sc)
	actual, _ := s.cache.LoadOrStore(sc, out)
	return actual.([]gworld.Landmark)
}

// computeLandmarks runs the full placement pipeline for one super-
// chunk. Steps:
//  1. Gate on anchor ocean — landmarks never spawn at sea.
//  2. Roll a count from the (60/30/8/2) distribution.
//  3. For each slot, attempt up to landmarkPlacementAttempts random
//     positions inside the super-chunk; reject ocean tiles, volcano
//     core/slope tiles, and tiles within landmarkMinSpacing of an
//     already-placed landmark.
//  4. Pick a kind from the region's character-affinity table.
//  5. Build a structured Parts name via naming.Generate.
//
// Pure: no goroutines, no shared mutation outside the sync.Map cache.
func (s *LandmarkSource) computeLandmarks(sc geom.SuperChunkCoord) []gworld.Landmark {
	anchor := geom.AnchorOf(s.seed, sc)
	if s.tileIsOcean(anchor) {
		return nil
	}

	// Independent PRNG stream for this super-chunk. Mixing seed with the
	// salt and the super-chunk coords yields a stable but decorrelated
	// stream — same (seed, sc) always produces the same draws regardless
	// of which order callers query super-chunks.
	state := uint64(s.seed) ^ uint64(landmarkSeedSalt) ^
		(uint64(int64(sc.X)) * 0x9e3779b97f4a7c15) ^
		(uint64(int64(sc.Y)) * 0xbf58476d1ce4e5b9)
	stream := geom.Splitmix64(state ^ 0x94d049bb133111eb)
	rng := rand.New(rand.NewPCG(state, stream))

	count := rollLandmarkCount(rng)
	if count == 0 {
		return nil
	}

	region := s.regions.RegionAt(sc)
	weights := kindWeightsFor(region.Character)

	out := make([]gworld.Landmark, 0, count)
	for slot := 0; slot < count; slot++ {
		pos, ok := s.pickPosition(rng, sc, out)
		if !ok {
			continue
		}
		kind := pickKind(rng, weights)
		name := naming.Generate(naming.Input{
			Domain:    naming.DomainLandmark,
			Character: region.Character.Key(),
			SubKind:   kind.Key(),
			Seed:      s.seed,
			CoordX:    pos.X,
			CoordY:    pos.Y,
		}, landmarkBounds())
		out = append(out, gworld.Landmark{
			Coord: pos,
			Kind:  kind,
			Name:  name,
		})
	}

	// Sort once for stable ordering across runs and cache reads. PCG
	// draws are already deterministic, but explicit lex sort future-
	// proofs against PCG implementation tweaks and makes test snapshots
	// readable in row-major order.
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Coord, out[j].Coord
		if a.Y != b.Y {
			return a.Y < b.Y
		}
		return a.X < b.X
	})
	return out
}

// rollLandmarkCount draws an integer in {0, 1, 2, 3} from the design-
// doc distribution. The thresholds are cumulative — a single uniform
// draw places the result in exactly one bucket.
func rollLandmarkCount(rng *rand.Rand) int {
	r := rng.Float64()
	switch {
	case r < landmarkProb0:
		return 0
	case r < landmarkProb0+landmarkProb1:
		return 1
	case r < landmarkProb0+landmarkProb1+landmarkProb2:
		return 2
	default:
		return 3
	}
}

// pickPosition tries up to landmarkPlacementAttempts random positions
// inside sc and returns the first one that clears every reject filter.
// Returns (zero, false) when every attempt fails — the caller drops
// the slot rather than retry forever.
func (s *LandmarkSource) pickPosition(
	rng *rand.Rand,
	sc geom.SuperChunkCoord,
	already []gworld.Landmark,
) (geom.Position, bool) {
	originX := sc.X * geom.SuperChunkSize
	originY := sc.Y * geom.SuperChunkSize

	for attempt := 0; attempt < landmarkPlacementAttempts; attempt++ {
		dx := rng.IntN(geom.SuperChunkSize)
		dy := rng.IntN(geom.SuperChunkSize)
		pos := geom.Position{X: originX + dx, Y: originY + dy}

		if s.tileIsOcean(pos) {
			continue
		}
		if s.tileInVolcanoZone(pos) {
			continue
		}
		if tooCloseToExisting(pos, already, landmarkMinSpacing) {
			continue
		}
		return pos, true
	}
	return geom.Position{}, false
}

// tileIsOcean reports whether the tile at p resolves to an ocean cell.
// Off-grid tiles read as ocean so a super-chunk straddling the world
// edge does not place landmarks in the void.
func (s *LandmarkSource) tileIsOcean(p geom.Position) bool {
	if s.world == nil || s.world.Voronoi == nil {
		return true
	}
	if p.X < 0 || p.Y < 0 || p.X >= s.world.Width || p.Y >= s.world.Height {
		return true
	}
	cellID := s.world.Voronoi.CellIDAt(p.X, p.Y)
	return s.world.IsOcean(cellID)
}

// tileInVolcanoZone reports whether p lies inside any volcano's core
// or slope ring. False when no VolcanoSource was wired or the index is
// empty.
func (s *LandmarkSource) tileInVolcanoZone(p geom.Position) bool {
	if s.volcanoZoneAt == nil {
		return false
	}
	_, exists := s.volcanoZoneAt[packPos(p)]
	return exists
}

// tooCloseToExisting returns true when p sits within minSpacing tiles
// (Chebyshev distance — max of |dx|, |dy|) of any landmark already
// placed in the same super-chunk. Chebyshev is cheaper than Euclidean
// and reads naturally on a square grid: "no two landmarks within an N-
// tile box".
func tooCloseToExisting(p geom.Position, existing []gworld.Landmark, minSpacing int) bool {
	for _, lm := range existing {
		dx := p.X - lm.Coord.X
		if dx < 0 {
			dx = -dx
		}
		dy := p.Y - lm.Coord.Y
		if dy < 0 {
			dy = -dy
		}
		if dx < minSpacing && dy < minSpacing {
			return true
		}
	}
	return false
}

// pickKind draws a LandmarkKind from a weighted distribution. The
// weights need not be normalized — the cumulative sum is computed at
// draw time so callers can express ratios directly.
func pickKind(rng *rand.Rand, weights []kindWeight) gworld.LandmarkKind {
	if len(weights) == 0 {
		return gworld.LandmarkTower
	}
	var total float32
	for _, w := range weights {
		total += w.Weight
	}
	if total <= 0 {
		return weights[0].Kind
	}
	r := rng.Float32() * total
	var cum float32
	for _, w := range weights {
		cum += w.Weight
		if r < cum {
			return w.Kind
		}
	}
	return weights[len(weights)-1].Kind
}

var _ gworld.LandmarkSource = (*LandmarkSource)(nil)
