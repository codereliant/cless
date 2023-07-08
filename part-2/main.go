package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

var containerManager ContainerManager

func main() {
	repo := &InMemoryServiceDefinitionRepository{
		services: make(map[string]ServiceDefinition),
		mutex:    &sync.Mutex{},
	}
	manager := NewServiceDefinitionManager(repo)
	go StartAdminServer(manager)
	containerManager = NewDockerContainerManager(manager)
	http.HandleFunc("/", handler)
	fmt.Println("Starting cless serverless reverse proxy server on port 80")
	http.ListenAndServe(":80", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// handle admin requests
	if r.Host == AdminHost {
		proxyToURL(w, r, fmt.Sprintf("%s:%d", "localhost", AdminPort))
		return
	}

	svcLocalHost, err := containerManager.GetRunningServiceForHost(r.Host)
	if err != nil {
		fmt.Printf("Failed to get running service: %s\n", err)
		w.Write([]byte("Failed to get running service"))
		return
	}
	fmt.Printf("Proxying to %s\n", *svcLocalHost)
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   *svcLocalHost,
	})
	proxy.ServeHTTP(w, r)
}

func proxyToURL(w http.ResponseWriter, r *http.Request, pURL string) {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   pURL,
	})
	proxy.ServeHTTP(w, r)
}
