package loadbalancer

import (
	"context"
	"fmt"
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

	t.Run("", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// Setup
		urls := []string{
			"http://localhost:8090",
			"http://localhost:8000",
			"http://localhost:3000",
		}

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

		// End Setup

		go func(ctx context.Context, services []*Service) {
			server := http.Server{Addr: ":8080", Handler: &RoundRobin{
				services: services,
			}}

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
			fmt.Println(string(b))
			time.Sleep(500 * time.Millisecond)
		}

	})

}
