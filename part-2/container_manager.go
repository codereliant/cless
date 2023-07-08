package main

import (
	"fmt"
	"math/rand"
	"net/http"
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

func (rSvc *RunningService) GetHost() string {
	return fmt.Sprintf("localhost:%d", rSvc.AssignedPort)
}

type ContainerManager interface {
	GetRunningServiceForHost(host string) (*string, error)
}

type DockerContainerManager struct {
	mutex       *sync.Mutex
	containers  map[string]*RunningService
	usedPorts   map[int]bool
	sDefManager *ServiceDefinitionManager
}

func NewDockerContainerManager(manager *ServiceDefinitionManager) *DockerContainerManager {
	return &DockerContainerManager{
		mutex:       &sync.Mutex{},
		containers:  make(map[string]*RunningService),
		usedPorts:   make(map[int]bool),
		sDefManager: manager,
	}
}

func (cm *DockerContainerManager) GetRunningServiceForHost(host string) (*string, error) {
	name := strings.Split(host, ".")[0]
	fmt.Printf("getting container for %s \n", name)
	sDef, err := cm.sDefManager.GetServiceDefinitionByName(name)
	fmt.Println("got service definition", sDef)
	if err != nil {
		return nil, err
	}
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	rSvc, exists := cm.containers[sDef.Name]
	if !exists {
		rSvc, err = cm.startContainer(sDef)
		if err != nil {
			fmt.Printf("Failed to start container: %s\n", err)
			return nil, err
		}
	}

	if !cm.isContainerReady(rSvc) {
		return nil, fmt.Errorf("container %s not ready", sDef.Name)
	}
	svcLocalHost := rSvc.GetHost()
	return &svcLocalHost, nil
}

func (cm *DockerContainerManager) startContainer(sDef *ServiceDefinition) (*RunningService, error) {
	fmt.Println("Starting container......")
	port := cm.getUnusedPort()
	fmt.Println("got port......")
	rSvc, err := cm.createContainer(sDef, port)
	if err != nil {
		return nil, err
	}
	cm.containers[sDef.Name] = rSvc
	cm.usedPorts[port] = true
	return rSvc, err
}

// create container with docker run
func (cm *DockerContainerManager) createContainer(sDef *ServiceDefinition, assignedPort int) (*RunningService, error) {
	image := fmt.Sprintf("%s:%s", sDef.ImageName, sDef.ImageTag)
	portMapping := fmt.Sprintf("%d:%d", assignedPort, sDef.Port)
	args := []string{"run", "-d"}
	args = append(args, "-p", portMapping)
	args = append(args, image)
	fmt.Println("docker", args)
	cmd := exec.Command("docker", args...)
	containerID, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to start container: %s\n", err)
		return nil, err
	}

	rSvc := RunningService{
		ContainerID:  string(containerID),
		AssignedPort: assignedPort,
		Ready:        false,
	}

	return &rSvc, nil
}

func (cm *DockerContainerManager) getUnusedPort() int {
	// get random port between 8000 and 9000
	// check if port is in use
	for {
		port := rand.Intn(1000) + 8000
		fmt.Println("checking port", port)
		_, exists := cm.usedPorts[port]
		if !exists {
			return port
		}
	}
}

func (cm *DockerContainerManager) isContainerReady(rSvc *RunningService) bool {
	if rSvc.Ready {
		return true
	}
	start := time.Now()
	for i := 0; i < 30; i++ {
		fmt.Println("Waiting for container to start...")
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", rSvc.AssignedPort))
		if err != nil {
			fmt.Println(err.Error())
		}
		if resp != nil && resp.StatusCode == 200 {
			fmt.Println("Container ready!")
			fmt.Printf("Container started in %s\n", time.Since(start))
			rSvc.Ready = true
			return true
		}
		fmt.Println("Container not ready yet...")
		time.Sleep(1 * time.Second)
	}
	return false
}
