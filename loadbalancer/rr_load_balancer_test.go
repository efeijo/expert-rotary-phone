package loadbalancer

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestRoundRobinLoadBalancer(t *testing.T) {

	t.Run("", func(t *testing.T) {
		//ctx, cancel := context.WithCancel(context.Background())
		// Setup
		urls := []string{
			"http://localhost:8090",
			"http://localhost:8000",
			"http://localhost:3000",
		}

		services := make([]*Service, 0, len(urls))

		for i, stringUrl := range urls {
			u, err := url.Parse(stringUrl)
			if err != nil {
				t.Error(err)
			}
			reverseProxy := &httputil.ReverseProxy{
				Rewrite: func(pr *httputil.ProxyRequest) {
					buf := bytes.NewBufferString(fmt.Sprintf("server %d", i))
					pr.Out.Write(buf)
				},
			}

			services = append(services, &Service{
				URL:          u,
				ReverseProxy: reverseProxy,
				Alive:        true,
			})
			go func(u *url.URL, reverseProxy *httputil.ReverseProxy) {
				log.Println(u.Host)
				log.Fatal(http.ListenAndServe(u.Host, reverseProxy))
			}(u, reverseProxy)
		}

		// End Setup

		go func(services []*Service) {
			log.Fatal(http.ListenAndServe(":8080", &RoundRobin{
				services: services,
			}))
		}(services)

		/* for i := 0; i < 6; i++ {

			resp, err := http.Get("http://localhost:8080")
			assert.NoError(t, err)
			b, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			fmt.Println(string(b))
			time.Sleep(500 * time.Millisecond)

		} */

		time.Sleep(30 * time.Second)
		os.Exit(1)
	})

}
