package render

import "fmt"

// RenderBracket generates a PNG bracket image at outputPath from the given match nodes.
//
// format must be one of "swiss", "single-elimination", or "double-elimination".
// name is used as the bracket title; pass an empty string to omit the title.
//
// Rendering takes 500ms–2s (headless Chrome). Call from a goroutine — do not
// block a Discord command handler on this.
func RenderBracket(nodes []MatchNode, format, name, outputPath string) error {
	html, err := buildHTML(nodes, format, name)
	if err != nil {
		return fmt.Errorf("building HTML: %w", err)
	}

	if err := screenshotHTML(html, outputPath); err != nil {
		return fmt.Errorf("screenshotting HTML: %w", err)
	}

	return nil
}
