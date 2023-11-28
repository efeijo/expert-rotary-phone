package loadbalancer

import (
	"context"
	"net/http"
)

type LoadBalancer interface {
	NextService() *Service
	PingServices(context.Context)
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}
