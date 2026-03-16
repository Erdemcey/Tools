package main

import (
	"bufio"
	"context"
	"fmt"
	"harbinger/scanner"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	cancel         context.CancelFunc
	intruderCancel context.CancelFunc
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

// --- REPEATER MODÜLÜ ---

func (a *App) SendRepeater(rawRequest string) string {
	rawRequest = strings.ReplaceAll(rawRequest, "\r\n", "\n")
	parts := strings.SplitN(rawRequest, "\n\n", 2)
	headerLines := strings.Split(parts[0], "\n")
	if len(headerLines) < 1 {
		return "Hata: Gecersiz Istek"
	}
	firstLine := strings.Fields(headerLines[0])
	if len(firstLine) < 2 {
		return "Hata: Method veya URL eksik"
	}

	method := strings.ToUpper(firstLine[0])
	targetURL := firstLine[1]
	var bodyReader io.Reader
	if len(parts) > 1 {
		bodyReader = strings.NewReader(parts[1])
	}
	req, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		return fmt.Sprintf("Hata: %s", err)
	}
	for _, line := range headerLines[1:] {
		h := strings.SplitN(line, ":", 2)
		if len(h) == 2 {
			req.Header.Set(strings.TrimSpace(h[0]), strings.TrimSpace(h[1]))
		}
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Baglanti Hatasi: %s", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("%s %s\n\n%s", resp.Proto, resp.Status, string(body))
}

// --- INTRUDER MODÜLÜ ---

func (a *App) StartIntruder(rawRequest string, payloadType string, manualPayload string, wordlistPath string, threads int) {
	if a.intruderCancel != nil {
		a.intruderCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.intruderCancel = cancel

	var payloads []string
	if payloadType == "manual" {
		payloads = strings.Split(manualPayload, "\n")
	} else {
		file, err := os.Open(wordlistPath)
		if err != nil {
			runtime.EventsEmit(a.ctx, "scan_complete", "Hata: Dosya bulunamadi.")
			return
		}
		s := bufio.NewScanner(file)
		for s.Scan() {
			payloads = append(payloads, s.Text())
		}
		file.Close()
	}

	jobs := make(chan string, threads)
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case p, ok := <-jobs:
					if !ok {
						return
					}
					finalReqStr := strings.ReplaceAll(rawRequest, "§", p)
					a.executeRawRequest(ctx, finalReqStr, p)
				}
			}
		}()
	}

	go func() {
		for _, p := range payloads {
			if strings.TrimSpace(p) == "" {
				continue
			}
			select {
			case <-ctx.Done():
				goto Finalize
			case jobs <- p:
			}
		}
	Finalize:
		close(jobs)
		wg.Wait()
		runtime.EventsEmit(a.ctx, "scan_complete", "Saldırı bitti veya durduruldu.")
	}()
}

func (a *App) StopIntruder() {
	if a.intruderCancel != nil {
		a.intruderCancel()
		a.intruderCancel = nil
	}
}

func (a *App) executeRawRequest(ctx context.Context, rawStr string, currentPayload string) {
	rawStr = strings.ReplaceAll(rawStr, "\r\n", "\n")
	parts := strings.SplitN(rawStr, "\n\n", 2)
	headerLines := strings.Split(parts[0], "\n")
	if len(headerLines) < 1 {
		return
	}
	firstLine := strings.Fields(headerLines[0])
	if len(firstLine) < 2 {
		return
	}
	method := strings.ToUpper(firstLine[0])
	targetURL := firstLine[1]
	var bodyReader io.Reader
	if len(parts) > 1 {
		bodyReader = strings.NewReader(parts[1])
	}
	req, _ := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	for _, line := range headerLines[1:] {
		h := strings.SplitN(line, ":", 2)
		if len(h) == 2 {
			req.Header.Set(strings.TrimSpace(h[0]), strings.TrimSpace(h[1]))
		}
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	runtime.EventsEmit(a.ctx, "found_result", map[string]interface{}{
		"Source":     "intruder",
		"StatusCode": resp.StatusCode,
		"URL":        currentPayload,
		"ContentLen": len(body),
	})
}
