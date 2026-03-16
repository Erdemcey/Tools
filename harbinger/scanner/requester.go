package scanner

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"
)

type Requester struct {
	Client *http.Client
}

func NewRequester(timeout time.Duration, threads int) *Requester {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// Bağlantı havuzu ayarları: ffuf benzeri hız için threads sayısına göre dinamik yapıyoruz
		MaxIdleConns:        0,           // Sınırsız toplam boşta bağlantı
		MaxIdleConnsPerHost: threads + 5, // Host başına boşta bekleyen bağlantı sayısı
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false, // Bağlantıları açık tutmak hız kazandırır

		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &Requester{
		Client: &http.Client{
			Transport: transport,
			Timeout:   timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Yönlendirmeleri takip etme
			},
		},
	}
}

func (r *Requester) DoRequest(url string) (int, int64, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) HarbingerScanner/1.0")

	resp, err := r.Client.Do(req)
	if err != nil {
		return 0, 0, err
	}

	// KRİTİK: Body'yi tamamen okumadan kapatırsan bağlantı reuse edilemez (Hız düşer)
	defer resp.Body.Close()
	n, _ := io.Copy(io.Discard, resp.Body)

	return resp.StatusCode, n, nil
}
