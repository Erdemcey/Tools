package scanner

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Requester, özelleştirilmiş HTTP istemcisini tutar
type Requester struct {
	Client *http.Client
}

// NewRequester, yüksek performans ayarlı bir istemci döner
func NewRequester(timeout time.Duration) *Requester {
	// Buradaki Transport ayarları hızın anahtarıdır
	transport := &http.Transport{
		// SSL sertifika hatalarını görmezden gel (Tarama araçları için kritik)
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},

		// Bağlantı havuzu ayarları
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,

		// TCP seviyesinde optimizasyonlar
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &Requester{
		Client: &http.Client{
			Transport: transport,
			Timeout:   timeout,
			// Otomatik yönlendirmeleri (Redirect) kapatmak genellikle daha iyidir
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// DoRequest, verilen URL'ye istek atar ve sonucu döner
func (r *Requester) DoRequest(url string) (int, int64, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, err
	}

	// WAF engellerine takılmamak için standart bir User-Agent ekleyelim
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) HarbingerScanner/1.0")

	resp, err := r.Client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// Content-Length (Boyut) bilgisini al
	size := resp.ContentLength

	return resp.StatusCode, size, nil
}
