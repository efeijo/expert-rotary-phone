package loadbalancer

import (
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type RoundRobin struct {
	services []*Service
	current  atomic.Uint32
}

// NextService implements LoadBalancer.
func (rr *RoundRobin) NextService() *Service {
	var service *Service
	idx := rr.current.Load()
	for {
		service = rr.services[int(idx%uint32(len(rr.services)))]

		if service.GetServiceStatus() {
			idx = rr.current.Add(1)
			return service
		}
		idx = rr.current.Add(1)
	}
}

func (rr *RoundRobin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	nextService := rr.NextService()
	if nextService == nil {
		log.Fatal("load balancer doesn't have any healthy services")
	}

	nextService.ReverseProxy.ServeHTTP(w, req)

}

func NewRoundRobinLoadBalancer(services []*Service) LoadBalancer {
	return &RoundRobin{
		services: services,
	}
}

func (rr *RoundRobin) PingServices() {
	for i, service := range rr.services {
		go func(i int, s *Service) {
			log.Println("pinging service:", s.URL.Host)
			err := backoff.Retry(
				func() error {
					conn, err := net.Dial("tcp", s.URL.Host)
					conn.Close()
					return err

				},
				backoff.WithMaxRetries(
					backoff.NewConstantBackOff(time.Second),
					3,
				),
			)
			rr.services[i].SetServiceStatus(err == nil)
		}(i, service)
	}
}
