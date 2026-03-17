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
		},
	}
}

// 🔥 BODY ve RAW AYRI DÖNÜYOR
func (r *Requester) DoRequest(url string) (int, int, string, string, error) {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "*/*")

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, 0, "", "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// --- RAW RESPONSE ---
	raw := fmt.Sprintf("%s %s\r\n", resp.Proto, resp.Status)

	for name, values := range resp.Header {
		for _, value := range values {
			raw += fmt.Sprintf("%s: %s\r\n", name, value)
		}
	}

	raw += "\r\n" + string(bodyBytes)

	return resp.StatusCode, len(bodyBytes), string(bodyBytes), raw, nil
}
