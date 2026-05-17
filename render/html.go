package render

import (
	"bytes"
	"fmt"
	"html/template"
)

// teamRowArgs is the data passed to the teamRow sub-template.
type teamRowArgs struct {
	Name    string
	Score   int
	Class   string
	Pending bool
}

// buildHTML selects and renders the appropriate HTML template for the given format.
func buildHTML(nodes []MatchNode, format, name string) (string, error) {
	switch format {
	case "swiss":
		return buildHTMLSwiss(computeSwissGrid(nodes, name))
	default: // "single-elimination", "double-elimination" (until double-elim gets its own template)
		return buildHTMLSingleElim(groupNodes(nodes, format, name))
	}
}

func buildHTMLSingleElim(b bracket) (string, error) {
	return renderTemplate("templates/bracket_single_elim.html", b)
}

func buildHTMLSwiss(b swissBracket) (string, error) {
	return renderTemplate("templates/bracket_swiss.html", b)
}

// renderTemplate loads a template from the embedded FS, executes it with data,
// and returns the resulting HTML string.
func renderTemplate(path string, data any) (string, error) {
	src, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading template %s: %w", path, err)
	}

	tmpl, err := template.New("bracket").Funcs(templateFuncs()).Parse(string(src))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %s: %w", path, err)
	}
	return buf.String(), nil
}

// templateFuncs returns the shared FuncMap used by all bracket templates.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"add":        func(a, b int) int { return a + b },
		"sub":        func(a, b int) int { return a - b },
		"div":        func(a, b int) int { return a / b },
		"not":        func(v bool) bool { return !v },
		"matchCount": func(r round) int { return len(r.Matches) },
		"makeRange": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"teamRowArgs": func(m match, teamName string, score int) teamRowArgs {
			cssClass := "pending"
			if !m.IsPending() {
				if teamName == m.Winner {
					cssClass = "winner"
				} else {
					cssClass = "loser"
				}
			}
			return teamRowArgs{
				Name:    teamName,
				Score:   score,
				Class:   cssClass,
				Pending: m.IsPending(),
			}
		},
	}
}
