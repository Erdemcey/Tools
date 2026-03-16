package scanner

import (
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
			// Redirect'leri takip etmemesi için (isteğe bağlı)
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// 4 değer döndürecek şekilde güncellendi: (status, size, body, error)
func (r *Requester) DoRequest(url string) (int, int, string, error) {
	resp, err := r.client.Get(url)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, 0, "", err
	}

	return resp.StatusCode, len(bodyBytes), string(bodyBytes), nil
}
