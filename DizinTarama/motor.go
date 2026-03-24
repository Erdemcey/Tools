package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8080/"
	directories := []string{"admin", "config", "images", "db", "uploads", "api"}

	// Özel bir istemci oluşturuyoruz (Zaman aşımı çok önemli!)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fmt.Printf("Tarama başlatılıyor: %s\n", baseURL)
	fmt.Println("---------------------------------")

	for _, dir := range directories {
		url := baseURL + dir

		// İstek oluşturma
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Go-Directory-Scanner-v1")

		// İsteği gönderme
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[!] %s -> Hata: %v\n", dir, err)
			continue
		}

		// Sonucu değerlendirme
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("[+] BULUNDU: %s (Kod: 200)\n", url)
		} else if resp.StatusCode == http.StatusForbidden {
			fmt.Printf("[*] YASAKLI: %s (Kod: 403)\n", url)
		}

		resp.Body.Close()
	}
}
