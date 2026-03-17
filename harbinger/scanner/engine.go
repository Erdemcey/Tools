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

// 🔥 DOĞRU YER: Body ve Raw burada
type Result struct {
	StatusCode int
	URL        string
	ContentLen int
	Method     string
	Body       string
	Raw        string
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
		Results:      make(chan Result, 1000),
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

	// --- WORKERS ---
	for i := 0; i < e.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for word := range words {
				select {
				case <-ctx.Done():
					return
				default:
					targetURL := fmt.Sprintf("%s/%s",
						strings.TrimRight(e.BaseURL, "/"),
						word,
					)

					// 🔥 DOĞRU: 5 değer al
					status, size, body, raw, err := reqClient.DoRequest(targetURL)

					if err == nil && status != 404 {
						select {
						case e.Results <- Result{
							URL:        targetURL,
							StatusCode: status,
							ContentLen: size,
							Method:     "GET",
							Body:       body, // sadece HTML
							Raw:        raw,  // full HTTP
						}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
	}

	// --- WORDLIST OKUMA ---
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			goto STOP
		case words <- scanner.Text():
		}
	}

STOP:
	close(words)
	wg.Wait()
}
