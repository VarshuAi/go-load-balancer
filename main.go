package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type ServerPool struct {
	backends []*url.URL
	current  uint64
}

func (s *ServerPool) AddBackend(u *url.URL) {
	s.backends = append(s.backends, u)
}

func (s *ServerPool) GetNextBackend() *url.URL {
	idx := atomic.AddUint64(&s.current, 1) % uint64(len(s.backends))
	return s.backends[idx]
}

func main() {
	port := flag.String("p", "8080", "Port to serve load balancer")
	flag.Parse()

	servers := []string{
		"http://127.0.0.1:8081",
		"http://127.0.0.1:8082",
	}

	serverPool := &ServerPool{}

	for _, s := range servers {
		u, err := url.Parse(s)
		if err != nil {
			log.Fatal(err)
		}
		serverPool.AddBackend(u)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := serverPool.GetNextBackend()
		fmt.Printf("[*] Load Balancer routing request -> Target Backend: %s\n", target)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	})

	fmt.Printf("[+] Load balancer running on port %s routing to backend instances: %v...\n", *port, servers)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}