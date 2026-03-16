package scanner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	ContentLen int64
}

type Engine struct {
	Threads      int
	WordlistPath string
	BaseURL      string
	Results      chan Result
}

func NewEngine(baseURL string, threads int, path string) *Engine {
	return &Engine{
		BaseURL:      baseURL,
		Threads:      threads,
		WordlistPath: path,
		Results:      make(chan Result, 1000), // Buffer'ı yüksek tutmak tıkanmayı önler
	}
}

func (e *Engine) Run(ctx context.Context) {
	defer close(e.Results)

	file, err := os.Open(e.WordlistPath)
	if err != nil {
		return
	}
	defer file.Close()

	words := make(chan string, e.Threads*10)
	var wg sync.WaitGroup
	reqClient := NewRequester(10*time.Second, e.Threads)

	// Worker Pool başlatılıyor
	for i := 0; i < e.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for word := range words {
				select {
				case <-ctx.Done(): // Stop'a basıldıysa hemen çık
					return
				default:
					targetURL := fmt.Sprintf("%s/%s", strings.TrimRight(e.BaseURL, "/"), word)

					status, size, err := reqClient.DoRequest(targetURL)

					if err == nil && status != 404 {
						select {
						case e.Results <- Result{
							URL:        targetURL,
							StatusCode: status,
							ContentLen: size,
						}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
	}

	// Dosyayı oku ve kanala gönder
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			goto StopProcessing
		case words <- scanner.Text():
		}
	}

StopProcessing:
	close(words) // Worker'lara iş bitti sinyali gönder
	wg.Wait()    // Tüm worker'ların güvenli kapandığından emin ol
}
