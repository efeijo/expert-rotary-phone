package loadbalancer

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type RoundRobin struct {
	services []*Service
	current  atomic.Uint32
}

func NewRoundRobinLoadBalancer(ctx context.Context, services []*Service) LoadBalancer {
	rr := &RoundRobin{services: services}
	go rr.PingServices(ctx)
	return rr
}

// NextService implements LoadBalancer.
func (rr *RoundRobin) NextService() *Service {
	var service *Service
	idx := rr.current.Load()
	for {
		circIndex := int(idx % uint32(len(rr.services)))
		service = rr.services[circIndex]
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

func (rr *RoundRobin) PingServices(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second).C
	for {
		select {
		case <-ticker:
			wg := &sync.WaitGroup{}
			for i, service := range rr.services {
				go func(wg *sync.WaitGroup, i int, s *Service) {
					defer wg.Done()
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
				}(wg, i, service)
			}
			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}
