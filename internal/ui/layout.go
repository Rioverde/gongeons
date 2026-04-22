package ui

// Constants that decide how much of the terminal the grid gets vs. the
// side panels and status bar. Tweak here if the box chrome changes.
const (
	// sidebarWidth is the fixed column budget for the right sidebar that
	// holds the stats panel. Wide enough for stat names + values + box
	// chrome, with a comfortable margin.
	sidebarWidth = 28

	// columnGap is the space between the left (map+events) column and the
	// right (stats) column. Two spaces keeps the layout breathable.
	columnGap = 2

	// gridBoxChrome is the total columns the grid's lipgloss box adds
	// around its tile content (border left+right plus horizontal padding).
	gridBoxChrome = 6

	// mapStatusRows is the row budget for the in-map status strip: two
	// info rows (region + server on row 1, coords + keybindings on row 2)
	// directly above the bottom border — no spacer, no divider.
	mapStatusRows = 2

	// mapHeaderRows is the row budget for the in-map date HUD header:
	// one row for the date line plus one horizontal rule between it
	// and the tile grid. Mirrors the pre-status horizontal rule so the
	// layout reads symmetrically above and below the grid.
	mapHeaderRows = 2

	// mapStatusDividerRows is the row budget for the horizontal rule
	// drawn between the tile grid and the status strip inside the map
	// box. One row. Split out from mapStatusRows so it can be accounted
	// for separately in the viewport-size calculation.
	mapStatusDividerRows = 1

	// gridBoxChromeV is the vertical chrome around the grid+status box.
	// renderMapBox uses Padding(0, 2) — zero vertical padding — so chrome
	// collapses to just the top and bottom borders.
	gridBoxChromeV = 2

	// eventsRows is the fixed row height of the events panel that sits
	// below the map in the left column. 5 rows = 2 chrome (top+bottom
	// border) + 3 visible event lines.
	eventsRows = 5

	// eventsBoxChromeV is the vertical chrome around the events box
	// (top border + padding-top + padding-bottom + bottom border).
	eventsBoxChromeV = 2

	// minTermWidth is the threshold below which we skip the sidebar and
	// stack everything vertically, matching the legacy layout.
	minTermWidth = 60

	// minViewportW / minViewportH match server's MinViewportWidth /
	// MinViewportHeight. If the terminal is smaller than this, the server
	// still clamps up, so we don't bother cropping here.
	minViewportW = 11
	minViewportH = 7

	// tileWidth is the number of terminal cells each map tile occupies
	// horizontally. Two cells per tile corrects the ~2:1 tall:wide aspect
	// ratio of a terminal cell so a game-tile reads as roughly square.
	tileWidth = 2
)

// TODO(Rioverde): add a runtime zoom toggle (+/- keys) that changes
// tileWidth or caps the viewport, letting the player trade tile density
// for field of view. For now the map fills the entire available space.

// viewportForTerm computes the widest × tallest odd-sided grid that fits
// inside the given terminal dimensions:
//
//   - The right sidebar (sidebarWidth + columnGap) is reserved horizontally
//     when the terminal is wide enough for the two-column layout.
//   - The events panel (eventsRows) sits below the map in the left column.
//   - The map box contains the tile grid and a one-row in-map status strip,
//     wrapped in a single lipgloss border (gridBoxChromeV accounts for the
//     border + padding rows).
//   - Each tile renders as tileWidth terminal cells wide (aspect-ratio fix).
//
// The map fills all remaining space — no DF-style cap. Odd sides keep the
// player glyph perfectly centred under the follow-camera.
func viewportForTerm(termW, termH int) (int, int) {
	horizReserved := 0
	if termW >= minTermWidth {
		horizReserved = sidebarWidth + columnGap
	}
	// Available terminal cells for tile content (horizontal).
	availCells := termW - horizReserved - gridBoxChrome
	// Each tile occupies tileWidth cells, so divide to get tile count.
	w := availCells / tileWidth
	// Vertical: full height minus map box chrome, the in-map date HUD
	// header (plus its divider), the in-map status strip and its
	// divider. The events panel lives below the map in the narrow
	// layout (so we still have to reserve rows for it there), but in
	// the wide layout it moves into the right column under Stats —
	// reclaiming those rows for the map grid so the viewport fills the
	// freed vertical space instead of leaving it blank beneath the map
	// box.
	h := termH - gridBoxChromeV -
		mapHeaderRows - mapStatusDividerRows - mapStatusRows
	if termW < minTermWidth {
		h -= eventsRows
	}

	if w < minViewportW {
		w = minViewportW
	}
	if h < minViewportH {
		h = minViewportH
	}
	if w%2 == 0 {
		w--
	}
	if h%2 == 0 {
		h--
	}
	return w, h
}
