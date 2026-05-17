package render

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// screenshotHTML renders htmlContent to a PNG file at outputPath.
//
// Both the rendered HTML and bracket.css are written to a temp directory so
// that the browser can resolve the relative stylesheet reference when loading
// via file:// URL. This also lets you open the source template in a browser
// for live CSS preview (the template dir has both files side-by-side).
func screenshotHTML(htmlContent, outputPath string) error {
	tmpDir, err := os.MkdirTemp("", "bracket-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "bracket.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("writing temp HTML: %w", err)
	}

	css, err := templateFS.ReadFile("templates/bracket.css")
	if err != nil {
		return fmt.Errorf("reading CSS: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "bracket.css"), css, 0644); err != nil {
		return fmt.Errorf("writing temp CSS: %w", err)
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(
		context.Background(),
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath("/usr/bin/chromium-browser"), // explicit path required on Fedora
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true), // required inside Docker
		)...,
	)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("file://"+htmlPath),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.FullScreenshot(&buf, 100),
	)
	if err != nil {
		return fmt.Errorf("chromedp run: %w", err)
	}

	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("writing PNG: %w", err)
	}
	return nil
}
