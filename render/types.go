package render

// MatchNode is the public input type. It mirrors the MatchNode in PickemsBot's
// store package so the bot can pass data directly without an adapter.
type MatchNode struct {
	ID      string // Liquipedia match2id, e.g. "IykJinz1G8_0001"
	Team1   string
	Team2   string
	Winner  string // team name, "TBD", or "" if unplayed
	Score   string // "2-1" (series) or "13-10" (BO1 map score); "" if unplayed
	Section string // round label, e.g. "Round 1", "Quarterfinals"
}

// The following types are internal to the render package.
// Fields are exported so html/template can access them via reflection.

type match struct {
	Team1  string
	Team2  string
	Score1 int
	Score2 int
	Winner string // empty string means pending
}

func (m match) IsPending() bool { return m.Winner == "" }

type round struct {
	Name    string
	Matches []match
}

type bracket struct {
	Name   string
	Rounds []round
}
