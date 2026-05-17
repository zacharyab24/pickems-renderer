package render

import (
	"sort"
	"strconv"
	"strings"
)

// knownSingleElimOrder maps Liquipedia section names to a fixed sort position
// for standard single-elimination bracket stages.
var knownSingleElimOrder = map[string]int{
	"Round of 64":   1,
	"Round of 32":   2,
	"Round of 16":   3,
	"Quarterfinals": 4,
	"Semifinals":    5,
	"Final":         6,
	"Grand Final":   6,
}

// groupNodes converts []MatchNode into a bracket by grouping by Section and
// sorting rounds according to the tournament format.
func groupNodes(nodes []MatchNode, format, name string) bracket {
	sectionMatches := map[string][]match{}
	var sectionOrder []string

	for _, n := range nodes {
		if _, seen := sectionMatches[n.Section]; !seen {
			sectionOrder = append(sectionOrder, n.Section)
		}
		sectionMatches[n.Section] = append(sectionMatches[n.Section], nodeToMatch(n))
	}

	sort.SliceStable(sectionOrder, func(i, j int) bool {
		return sectionRank(sectionOrder[i], format) < sectionRank(sectionOrder[j], format)
	})

	rounds := make([]round, 0, len(sectionOrder))
	for _, s := range sectionOrder {
		rounds = append(rounds, round{Name: s, Matches: sectionMatches[s]})
	}

	return bracket{Name: name, Rounds: rounds}
}

// sectionRank returns a numeric sort key for a section name given the format.
func sectionRank(section, format string) int {
	switch format {
	case "swiss":
		return parseTrailingInt(section)
	case "single-elimination":
		if rank, ok := knownSingleElimOrder[section]; ok {
			return rank
		}
		return parseTrailingInt(section)
	case "double-elimination":
		return doubleElimRank(section)
	default:
		return parseTrailingInt(section)
	}
}

// parseTrailingInt extracts the trailing integer from a string like "Round 3".
// Returns 0 if no integer is found.
func parseTrailingInt(s string) int {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0
	}
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0
	}
	return n
}

// doubleElimRank interleaves Upper and Lower bracket rounds by round number,
// placing Upper before Lower within the same round. Grand Final sorts last.
func doubleElimRank(section string) int {
	lower := strings.ToLower(section)
	if strings.Contains(lower, "grand final") {
		return 10000
	}
	base := parseTrailingInt(section) * 10
	if strings.HasPrefix(lower, "lower") {
		base++ // Lower sorts after Upper within the same round number
	}
	return base
}

// nodeToMatch converts a MatchNode to the internal match type.
// Score string "2-1" is split into Score1=2, Score2=1.
// Winner "TBD" or "" is normalised to "" (pending).
func nodeToMatch(n MatchNode) match {
	winner := n.Winner
	if winner == "TBD" {
		winner = ""
	}
	score1, score2 := parseScore(n.Score)
	return match{
		Team1:  n.Team1,
		Team2:  n.Team2,
		Score1: score1,
		Score2: score2,
		Winner: winner,
	}
}

// parseScore splits a score string like "2-1" into (2, 1).
// Returns (0, 0) for empty or malformed strings.
func parseScore(s string) (int, int) {
	if s == "" {
		return 0, 0
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	a, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	b, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return a, b
}
