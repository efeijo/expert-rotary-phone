package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Service struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	Alive        bool
	mu           sync.Mutex
}

func (s *Service) SetServiceStatus(alive bool) {
	s.mu.Lock()
	s.Alive = alive
	s.mu.Unlock()
}

func (s *Service) GetServiceStatus() bool {
	s.mu.Lock()
	alive := s.Alive
	s.mu.Unlock()
	return alive
}
