package loadbalancer

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRoundRobinLoadBalancer(t *testing.T) {
	urls := []string{
		"http://localhost:8090",
		"http://localhost:8000",
		"http://localhost:3000",
	}

	t.Run("should call each service once", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		//shutdow all servers
		defer cancel()
		// Setup
		services := make([]*Service, 0, len(urls))

		for _, stringUrl := range urls {
			u, err := url.Parse(stringUrl)
			if err != nil {
				t.Error(err)
			}
			reverseProxy := httputil.NewSingleHostReverseProxy(u)

			services = append(services, &Service{
				URL:          u,
				ReverseProxy: reverseProxy,
				Alive:        true,
			})
			go func(ctx context.Context, u *url.URL, reverseProxy *httputil.ReverseProxy) {
				server := http.Server{Addr: u.Host, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(u.String()))
				})}
				go func() {
					log.Fatal(server.ListenAndServe())
				}()

				<-ctx.Done()
				server.Shutdown(ctx)
			}(ctx, u, reverseProxy)
		}

		lb := NewRoundRobinLoadBalancer(ctx, services)
		// End Setup

		go func(ctx context.Context, services []*Service) {
			server := http.Server{Addr: ":8080", Handler: lb}

			go func() {
				log.Fatal(server.ListenAndServe())
			}()

			<-ctx.Done()
			server.Shutdown(ctx)
		}(ctx, services)

		for i := 0; i < 3; i++ {
			resp, err := http.Get("http://localhost:8080")
			assert.NoError(t, err)

			b, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			assert.Equal(t, urls[i], string(b))
			time.Sleep(1500 * time.Millisecond)
		}
	})
	t.Run("should call each service once", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		//shutdow all servers
		defer cancel()
		// Setup
		services := make([]*Service, 0, len(urls))

		for _, stringUrl := range urls {
			u, err := url.Parse(stringUrl)
			if err != nil {
				t.Error(err)
			}
			reverseProxy := httputil.NewSingleHostReverseProxy(u)

			services = append(services, &Service{
				URL:          u,
				ReverseProxy: reverseProxy,
				Alive:        true,
			})
			go func(ctx context.Context, u *url.URL, reverseProxy *httputil.ReverseProxy) {
				server := http.Server{Addr: u.Host, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(u.String()))
				})}
				go func() {
					log.Fatal(server.ListenAndServe())
				}()

				<-ctx.Done()
				server.Shutdown(ctx)
			}(ctx, u, reverseProxy)
		}

		lb := NewRoundRobinLoadBalancer(ctx, services)
		// End Setup

		go func(ctx context.Context, services []*Service) {
			server := http.Server{Addr: ":8080", Handler: lb}

			go func() {
				log.Fatal(server.ListenAndServe())
			}()

			<-ctx.Done()
			server.Shutdown(ctx)
		}(ctx, services)

		for i := 0; i < 3; i++ {
			resp, err := http.Get("http://localhost:8080")
			assert.NoError(t, err)

			b, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			assert.Equal(t, urls[i], string(b))
			time.Sleep(1500 * time.Millisecond)
		}
	})

}
