package render_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zacharyab24/pickems-renderer/render"
)

const outputDir = "testdata/output"

func setup(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("creating output dir: %v", err)
	}
}

func loadNodes(t *testing.T, name string) []render.MatchNode {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata/input", name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	var nodes []render.MatchNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		t.Fatalf("parsing %s: %v", name, err)
	}
	return nodes
}

func assertFileWritten(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("output file not found at %s: %v", path, err)
	}
	t.Logf("written → %s", path)
}

func TestRenderSwiss(t *testing.T) {
	setup(t)
	nodes := loadNodes(t, "swiss.json")
	out := filepath.Join(outputDir, "swiss.png")
	if err := render.RenderBracket(nodes, "swiss", "PGL Masters Bucharest 2025 Major — Group Stage", out); err != nil {
		t.Fatalf("RenderBracket: %v", err)
	}
	assertFileWritten(t, out)
}

func TestRenderSingleElim(t *testing.T) {
	setup(t)
	nodes := loadNodes(t, "single_elim.json")
	out := filepath.Join(outputDir, "single_elim.png")
	if err := render.RenderBracket(nodes, "single-elimination", "PGL Masters Bucharest 2025 Major — Playoffs", out); err != nil {
		t.Fatalf("RenderBracket: %v", err)
	}
	assertFileWritten(t, out)
}
