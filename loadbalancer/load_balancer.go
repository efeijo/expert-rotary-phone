package loadbalancer

type LoadBalancer interface {
	NextService() *Service
}
