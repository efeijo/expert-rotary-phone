package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	u, _ := url.Parse("http://localhost:8080")
	rp := httputil.NewSingleHostReverseProxy(u)

	go func() {
		http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hey foscasse"))
		}))
	}()
	http.ListenAndServe(":9090", rp)

}
