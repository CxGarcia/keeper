package proxy

import "net/http"

// Service defines the proxy service
type Service struct {
	httpClient *http.Client
}

// NewService creates a new proxy service
func NewService(httpClient *http.Client) *Service {
	return &Service{
		httpClient: httpClient,
	}
}

// ProxyRequest proxies the request to the target server
func (s *Service) ProxyRequest(req *http.Request) (*http.Response, error) {
	return s.httpClient.Do(req)
}
