package main

import (
	"bufio"
	"context"
	"fmt"
	"harbinger/scanner" // Modül adın farklıysa burayı güncelle
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	cancel         context.CancelFunc // Scanner için
	intruderCancel context.CancelFunc // Intruder için
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- SCANNER MODÜLÜ ---

func (a *App) StartScan(targetURL string, threads int, wordlistPath string) {
	if a.cancel != nil {
		a.cancel()
	}

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel

	engine := scanner.NewEngine(targetURL, threads, wordlistPath)

	go func() {
		engine.Run(ctx)
		for res := range engine.Results {
			runtime.EventsEmit(a.ctx, "found_result", res)
		}
		runtime.EventsEmit(a.ctx, "scan_complete", "Tarama tamamlandı.")
	}()
}

func (a *App) StopScan() {
	if a.cancel != nil {
		a.cancel()
	}
}

// --- REPEATER MODÜLÜ ---

func (a *App) SendRepeater(rawRequest string) string {
	reader := bufio.NewReader(strings.NewReader(rawRequest))

	// İlk satır: METHOD URL PROTOCOL
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return "Hata: İstek okunamadı."
	}

	parts := strings.Fields(firstLine)
	if len(parts) < 2 {
		return "Geçersiz Format: METHOD URL eksik."
	}

	method := parts[0]
	targetURL := parts[1]

	req, err := http.NewRequest(method, targetURL, nil)
	if err != nil {
		return fmt.Sprintf("Hata: %s", err)
	}

	// Headerları işle
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		h := strings.SplitN(line, ":", 2)
		if len(h) == 2 {
			req.Header.Set(strings.TrimSpace(h[0]), strings.TrimSpace(h[1]))
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Hata: %s", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("%s %s\n\n%s", resp.Proto, resp.Status, string(body))
}

// --- INTRUDER MODÜLÜ ---

func (a *App) StartIntruder(rawRequest string, payloadType string, manualPayload string, wordlistPath string, threads int) {
	// Eğer çalışan bir saldırı varsa iptal et
	if a.intruderCancel != nil {
		a.intruderCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.intruderCancel = cancel

	var payloads []string
	if payloadType == "manual" {
		payloads = append(payloads, manualPayload)
	} else {
		file, err := os.Open(wordlistPath)
		if err != nil {
			runtime.EventsEmit(a.ctx, "scan_complete", "Hata: Wordlist açılamadı.")
			return
		}
		scannerFile := bufio.NewScanner(file)
		for scannerFile.Scan() {
			payloads = append(payloads, scannerFile.Text())
		}
		file.Close()
	}

	jobs := make(chan string, len(payloads))
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					finalReqStr := strings.ReplaceAll(rawRequest, "§", p)
					a.executeRawRequest(finalReqStr, p)
				}
			}
		}()
	}

	for _, p := range payloads {
		jobs <- p
	}
	close(jobs)
	wg.Wait()
	runtime.EventsEmit(a.ctx, "scan_complete", "Saldırı Tamamlandı!")
}

// executeRawRequest: Intruder için ham isteği işleyip sonuç döner
func (a *App) executeRawRequest(rawStr string, currentPayload string) {
	// Satır sonlarını standardize et (Burp'ten gelen \r\n sorununu çözer)
	rawStr = strings.ReplaceAll(rawStr, "\r\n", "\n")

	// Body ve Header ayrımı
	parts := strings.SplitN(rawStr, "\n\n", 2)
	headerPart := parts[0]
	bodyPart := ""
	if len(parts) > 1 {
		bodyPart = parts[1]
	}

	// İlk satırı oku
	reader := bufio.NewReader(strings.NewReader(headerPart))
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	fParts := strings.Fields(firstLine)
	if len(fParts) < 2 {
		return
	}

	// Değişkenleri açıkça tanımlayalım
	method := strings.ToUpper(fParts[0])
	targetURL := fParts[1]

	// Body'yi hazırla
	var bodyReader io.Reader
	if bodyPart != "" {
		bodyReader = strings.NewReader(bodyPart)
	} else {
		bodyReader = nil // Body yoksa nil olmalı
	}

	// İsteği oluştur
	req, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		fmt.Println("Request Error:", err)
		return
	}

	// Headerları ekle (döngü devam ediyor...)
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		h := strings.SplitN(line, ":", 2)
		if len(h) == 2 {
			req.Header.Set(strings.TrimSpace(h[0]), strings.TrimSpace(h[1]))
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	runtime.EventsEmit(a.ctx, "found_result", map[string]interface{}{
		"StatusCode": resp.StatusCode,
		"URL":        currentPayload,
		"ContentLen": len(body),
		"Method":     method,
	})
}

// --- GENEL ARAÇLAR ---

func (a *App) SelectWordlist() string {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Wordlist Seç",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}
