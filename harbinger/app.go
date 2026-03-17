package main

import (
	"context"
	"fmt"
	"harbinger/scanner"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- GENEL ARAÇLAR ---
func (a *App) SelectWordlist() string {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Wordlist Sec",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

// --- SCANNER MODÜLÜ ---
func (a *App) StartScan(targetURL string, threads int, wordlistPath string) {
	a.StopScan()
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "http://" + targetURL
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	engine := scanner.NewEngine(targetURL, threads, wordlistPath)
	go engine.Run(ctx)

	go func() {
		for {
			select {
			case res, ok := <-engine.Results:
				if !ok {
					runtime.EventsEmit(a.ctx, "scan_complete", "Tarama Bitti")
					return
				}
				runtime.EventsEmit(a.ctx, "found_result", map[string]interface{}{
					"Source":     "scanner",
					"StatusCode": res.StatusCode,
					"URL":        res.URL,
					"ContentLen": res.ContentLen,
					"Method":     res.Method,
					"Body":       res.Body,
					"Raw":        res.Raw,
					"RawRequest": fmt.Sprintf("%s %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: Harbinger/1.0\r\nAccept: */*\r\n\r\n", res.Method, res.URL, targetURL),
				})
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (a *App) StopScan() {
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
		runtime.EventsEmit(a.ctx, "scan_complete", "Tarama durduruldu.")
	}
}
