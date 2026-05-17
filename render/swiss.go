package render

import (
	"fmt"
	"sort"
)

// subRows is the LCM of 1, 2, 3, 4 — ensures clean integer row spans for every
// possible column size in a standard Swiss bracket (max 4 cells per column).
const subRows = 12

type swissCell struct {
	GridColumn   int // css grid-column (1-indexed, = wins+losses+1)
	GridRowStart int // css grid-row start (1-indexed)
	GridRowEnd   int // css grid-row end (exclusive)
	Wins         int
	Losses       int
	Matches      []match
	Teams        []string // populated for qualify/eliminate terminal cells
	State        string   // "qualify", "eliminate", or ""
	Label        string   // "W:L" e.g. "2:1"
}

type swissBracket struct {
	Name       string
	Cells      []swissCell
	NumColumns int // total grid columns (= max rounds played + 1)
}

// computeSwissGrid derives the Swiss bracket layout from match nodes.
//
// Columns are keyed by rounds played (wins+losses):
//   col 1 → 0:0   col 2 → 1:0/0:1   col 3 → 2:0/1:1/0:2   col 4 → 3:0/2:1/1:2/0:3 …
//
// Row positions use a 12-sub-row grid (LCM of 1–4) so cells in every column get
// clean integer spans regardless of column size.
//
// Terminal cells (wins==3 or losses==3) are placed in the same column as the
// matches that determined them and rendered as team lists rather than match cards.
func computeSwissGrid(nodes []MatchNode, name string) swissBracket {
	// Group nodes by section (round), preserving insertion order.
	sectionMatches := map[string][]MatchNode{}
	var sectionOrder []string
	for _, n := range nodes {
		if _, seen := sectionMatches[n.Section]; !seen {
			sectionOrder = append(sectionOrder, n.Section)
		}
		sectionMatches[n.Section] = append(sectionMatches[n.Section], n)
	}
	sort.Slice(sectionOrder, func(i, j int) bool {
		return parseTrailingInt(sectionOrder[i]) < parseTrailingInt(sectionOrder[j])
	})

	// Initialise every real team at 0-0.
	type record struct{ wins, losses int }
	records := map[string]record{}
	for _, n := range nodes {
		for _, team := range []string{n.Team1, n.Team2} {
			if team != "" && team != "TBD" {
				if _, ok := records[team]; !ok {
					records[team] = record{}
				}
			}
		}
	}

	type cellKey struct{ wins, losses int }
	cellMap := map[cellKey][]match{}

	for _, section := range sectionOrder {
		// Snapshot records before this round — Swiss rounds are simultaneous.
		pre := make(map[string]record, len(records))
		for k, v := range records {
			pre[k] = v
		}

		for _, n := range sectionMatches[section] {
			if n.Team1 == "TBD" || n.Team2 == "TBD" {
				continue
			}
			r := pre[n.Team1]
			key := cellKey{r.wins, r.losses}
			cellMap[key] = append(cellMap[key], nodeToMatch(n))
		}

		for _, n := range sectionMatches[section] {
			if n.Winner == "" || n.Winner == "TBD" {
				continue
			}
			loser := n.Team2
			if n.Winner == n.Team2 {
				loser = n.Team1
			}
			w := records[n.Winner]
			w.wins++
			records[n.Winner] = w
			l := records[loser]
			l.losses++
			records[loser] = l
		}
	}

	// Collect match cells.
	prePositioned := map[cellKey]swissCell{}
	for key, matches := range cellMap {
		prePositioned[key] = swissCell{
			Wins:    key.wins,
			Losses:  key.losses,
			Matches: matches,
			Label:   fmt.Sprintf("%d:%d", key.wins, key.losses),
		}
	}

	// 3:0 and 0:3 stay as individual cells; 3:1/3:2 and 1:3/2:3 merge into aggregates.
	var qualifyTeams, eliminateTeams []string
	individualTerminals := map[cellKey][]string{}
	for team, rec := range records {
		key := cellKey{rec.wins, rec.losses}
		switch {
		case rec.wins == 3 && rec.losses == 0:
			individualTerminals[key] = append(individualTerminals[key], team)
		case rec.losses == 3 && rec.wins == 0:
			individualTerminals[key] = append(individualTerminals[key], team)
		case rec.wins == 3:
			qualifyTeams = append(qualifyTeams, team)
		case rec.losses == 3:
			eliminateTeams = append(eliminateTeams, team)
		}
	}
	for key, teams := range individualTerminals {
		sort.Strings(teams)
		state := "qualify"
		if key.losses == 3 {
			state = "eliminate"
		}
		prePositioned[key] = swissCell{
			Wins:   key.wins,
			Losses: key.losses,
			Teams:  teams,
			State:  state,
			Label:  fmt.Sprintf("%d:%d", key.wins, key.losses),
		}
	}
	sort.Strings(qualifyTeams)
	sort.Strings(eliminateTeams)

	// Group match cells into columns by rounds played (wins+losses), find max.
	colCells := map[int][]swissCell{}
	maxRounds := 0
	for _, c := range prePositioned {
		r := c.Wins + c.Losses
		colCells[r] = append(colCells[r], c)
		if r > maxRounds {
			maxRounds = r
		}
	}

	// Aggregate cells share the last match column. Sentinel Losses values
	// (-1 / maxInt) force them to the top and bottom when that column is sorted.
	if len(qualifyTeams) > 0 {
		colCells[maxRounds] = append(colCells[maxRounds], swissCell{
			Losses: -1,
			Teams:  qualifyTeams,
			State:  "qualify",
			Label:  "Advanced",
		})
	}
	if len(eliminateTeams) > 0 {
		colCells[maxRounds] = append(colCells[maxRounds], swissCell{
			Losses: 1<<31 - 1,
			Teams:  eliminateTeams,
			State:  "eliminate",
			Label:  "Eliminated",
		})
	}

	// Assign CSS grid positions.
	// Within each column, sort by losses ascending (fewest losses at top).
	// Row span: for a column of N cells, each cell gets subRows/N rows.
	var finalCells []swissCell
	numColumns := 0
	for r := 0; r <= maxRounds; r++ {
		cells := colCells[r]
		if len(cells) == 0 {
			continue
		}
		sort.Slice(cells, func(i, j int) bool {
			return cells[i].Losses < cells[j].Losses
		})
		n := len(cells)
		gridCol := r + 1
		if gridCol > numColumns {
			numColumns = gridCol
		}
		for i, c := range cells {
			c.GridColumn = gridCol
			c.GridRowStart = i*subRows/n + 1
			c.GridRowEnd = (i+1)*subRows/n + 1
			finalCells = append(finalCells, c)
		}
	}

	return swissBracket{
		Name:       name,
		Cells:      finalCells,
		NumColumns: numColumns,
	}
}
