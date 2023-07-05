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

type Container struct {
	ID   string
	Port int
}

const defaultPort = 8080

// We will use a map and a mutex to store and manage our docker containers
var mutex = &sync.Mutex{}
var containers = make(map[string]Container)
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
		containerID, port := startContainer(containerName)
		mutex.Lock()
		containers[containerName] = Container{
			ID:   containerID,
			Port: port,
		}
		mutex.Unlock()
	}

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", containers[containerName].Port),
	})
	proxy.ServeHTTP(w, r)
}

func startContainer(containerName string) (string, int) {
	port := getUnusedPort(containerName)
	fmt.Println("docker", "run", "-d", "-p", fmt.Sprintf("%d:%d", port, defaultPort), containerName)
	cmd := exec.Command("docker", "run", "-d", "-p", fmt.Sprintf("%d:%d", port, defaultPort), containerName)
	containerID, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to start container: %s\n", err)
		panic(err)
	}

	start := time.Now()
	for i := 0; i < 15; i++ {
		fmt.Println("Waiting for container to start...")
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
		if err != nil {
			fmt.Println(err.Error())
		}
		if resp != nil && resp.StatusCode == 200 {
			fmt.Println("Container ready!")
			break
		}
		fmt.Println("Container not ready yet...")
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("Container started in %s\n", time.Since(start))

	portToContainerID[port] = string(containerID)

	return string(containerID), port
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
