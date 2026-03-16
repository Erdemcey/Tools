package scanner

import (
	"bufio" // Satır satır okuma için
	"context"
	"fmt"
	"os" // Dosya işlemleri için
	"sync"
	"time"
)

// BU KISMI EKLE VEYA VARSA KONTROL ET
type Result struct {
	URL        string
	StatusCode int
	ContentLen int64
}

type Engine struct {
	Threads      int
	WordlistPath string // Artık liste değil, dosya yolu tutuyoruz
	BaseURL      string
	Results      chan Result
	WorkerPool   chan struct{}
}

func NewEngine(baseURL string, threads int, path string) *Engine {
	return &Engine{
		BaseURL:      baseURL,
		Threads:      threads,
		WordlistPath: path,
		Results:      make(chan Result),
		WorkerPool:   make(chan struct{}, threads),
	}
}

func (e *Engine) Run(ctx context.Context) {
	var wg sync.WaitGroup
	reqClient := NewRequester(10 * time.Second)

	// Dosyayı aç
	file, err := os.Open(e.WordlistPath)
	if err != nil {
		return // Hata durumunda çık
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Dosyadaki her satırı (kelimeyi) oku
	for scanner.Scan() {
		word := scanner.Text()

		select {
		case <-ctx.Done():
			return
		case e.WorkerPool <- struct{}{}:
			wg.Add(1)
			targetURL := fmt.Sprintf("%s/%s", e.BaseURL, word)

			go func(url string) {
				defer wg.Done()
				defer func() { <-e.WorkerPool }()

				status, size, err := reqClient.DoRequest(url)
				if err != nil {
					return
				}

				e.Results <- Result{
					URL:        url,
					StatusCode: status,
					ContentLen: size,
				}
			}(targetURL)
		}
	}

	go func() {
		wg.Wait()
		close(e.Results)
	}()
}
