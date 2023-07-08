package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type RunningService struct {
	ContainerID  string // docker container ID
	AssignedPort int    // port assigned to the container
	Ready        bool   // whether the container is ready to serve requests
}

const defaultPort = 8080 // let's assume that all the images expose port 8080

// We will use a map and a mutex to store and manage our docker containers
var mutex = &sync.Mutex{}

var containers = make(map[string]RunningService) // map of hostname to running services
var portToContainerID = make(map[int]string)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	containerName := getContainerNameFromHost(r.Host)
	mutex.Lock()
	_, exists := containers[containerName]
	mutex.Unlock()

	if !exists {
		mutex.Lock()
		rSvc, err := startContainer(containerName)
		if err != nil {
			fmt.Printf("Failed to start container: %s\n", err)
			w.Write([]byte("Failed to start container"))
			return
		}
		containers[containerName] = RunningService{
			ContainerID:  rSvc.ContainerID,
			AssignedPort: rSvc.AssignedPort,
		}
		mutex.Unlock()
		if !isContainerReady(*rSvc) {
			w.Write([]byte("Container not ready after 30 seconds"))
		} else {
			mutex.Lock()
			rSvc := containers[containerName]
			rSvc.Ready = true
			mutex.Unlock()
		}
	} else {
		mutex.Lock()
		rSvc := containers[containerName]
		mutex.Unlock()
		if !rSvc.Ready {
			w.Write([]byte("Container not ready yet"))
			return
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", containers[containerName].AssignedPort),
	})
	proxy.ServeHTTP(w, r)
}

func startContainer(containerName string) (*RunningService, error) {
	port := getUnusedPort(containerName)
	cmd := exec.Command("docker", "run", "-d", "-p", fmt.Sprintf("%d:%d", port, defaultPort), containerName)
	containerID, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to start container: %s\n", err)
		return nil, err
	}
	portToContainerID[port] = string(containerID)
	rSvc := RunningService{
		ContainerID:  string(containerID),
		AssignedPort: port,
		Ready:        false,
	}

	return &rSvc, nil
}

func getContainerNameFromHost(host string) string {
	parts := strings.Split(host, ".")
	return parts[0] + "-docker"
}

func getUnusedPort(containerName string) int {
	// get random port between 8000 and 9000
	// check if port is in use
	port := rand.Intn(1000) + 8000
	_, exists := portToContainerID[port]
	if exists {
		return getUnusedPort(containerName)
	}
	return port
}

func isContainerReady(rSvc RunningService) bool {
	start := time.Now()
	for i := 0; i < 29; i++ {
		fmt.Println("Waiting for container to start...")
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", rSvc.AssignedPort))
		if err != nil {
			fmt.Println(err.Error())
		}
		if resp != nil && resp.StatusCode == 200 {
			fmt.Println("Container ready!")
			fmt.Printf("Container started in %s\n", time.Since(start))
			return true
		}
		fmt.Println("Container not ready yet...")
		time.Sleep(1 * time.Second)
	}
	return false
}
