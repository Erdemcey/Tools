package scanner

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Requester struct {
	client *http.Client
}

func NewRequester(timeout time.Duration, threads int) *Requester {
	return &Requester{
		client: &http.Client{
			Timeout: timeout,
			// Burp Suite gibi davran: Yönlendirmeleri otomatik takip etme,
			// her adımın RAW cevabını görmemizi sağla.
			CheckRedirect: nil,
		},
	}
}

func (r *Requester) DoRequest(url string) (int, int, string, error) {
	req, _ := http.NewRequest("GET", url, nil)

	// Tarayıcı gibi görünmek için User-Agent eklemek çok kritiktir
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// --- TAM RAW RESPONSE OLUŞTURMA ---
	// Status Line: HTTP/1.1 200 OK
	raw := fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status)

	// Headers
	for name, values := range resp.Header {
		for _, value := range values {
			raw += fmt.Sprintf("%s: %s\r\n", name, value)
		}
	}

	// Body ile birleştir
	raw += "\r\n" + string(bodyBytes)

	return resp.StatusCode, len(bodyBytes), raw, nil
}
