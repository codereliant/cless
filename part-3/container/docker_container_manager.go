package container

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"codereliant.io/cless/admin"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
)

type DockerContainerManager struct {
	mutex        *sync.Mutex
	containers   map[string]*RunningService
	usedPorts    map[int]bool
	sDefManager  *admin.ServiceDefinitionManager
	dockerClient *client.Client
}

func NewDockerContainerManager(manager *admin.ServiceDefinitionManager) (ContainerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	mgr := &DockerContainerManager{
		mutex:        &sync.Mutex{},
		containers:   make(map[string]*RunningService),
		usedPorts:    make(map[int]bool),
		sDefManager:  manager,
		dockerClient: cli,
	}

	return mgr, nil
}

func (cm *DockerContainerManager) GetRunningServiceForHost(host string) (*string, error) {
	log.Debug().Str("host", host).Msg("getting container")
	sDef, err := cm.sDefManager.GetServiceDefinitionByHost(host)
	log.Debug().Str("service definition", host).Msg("got service definition")
	if err != nil {
		return nil, err
	}
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	rSvc, exists := cm.containers[sDef.Name]
	if !exists {
		rSvc, err = cm.startContainer(sDef)
		if err != nil {
			return nil, err
		}
	}

	if !cm.isContainerReady(rSvc) {
		return nil, fmt.Errorf("container %s not ready", sDef.Name)
	}
	svcLocalHost := rSvc.GetHost()
	return &svcLocalHost, nil
}

func (cm *DockerContainerManager) startContainer(sDef *admin.ServiceDefinition) (*RunningService, error) {
	port := cm.getUnusedPort()
	rSvc, err := cm.createContainer(sDef, port)
	if err != nil {
		return nil, err
	}
	cm.containers[sDef.Name] = rSvc
	cm.usedPorts[port] = true
	return rSvc, err
}

// create container with docker run
func (cm *DockerContainerManager) createContainer(sDef *admin.ServiceDefinition, assignedPort int) (*RunningService, error) {

	image := fmt.Sprintf("%s:%s", sDef.ImageName, sDef.ImageTag)
	ctx := context.Background()
	resp, err := cm.dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image: image,
			Tty:   false,
		},
		&container.HostConfig{
			PortBindings: buildPortBindings(sDef.Port, assignedPort),
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		return nil, err
	}

	if err := cm.dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	rSvc := RunningService{
		ContainerID:  string(resp.ID),
		AssignedPort: assignedPort,
		Ready:        false,
	}

	return &rSvc, nil
}

func buildPortBindings(sDefPort, assignedPort int) nat.PortMap {
	portBindings := nat.PortMap{
		nat.Port(fmt.Sprintf("%d/tcp", sDefPort)): []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: fmt.Sprintf("%d", assignedPort),
			},
		},
	}

	return portBindings
}

func (cm *DockerContainerManager) StopAndRemoveAllContainers() []error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	var errors []error
	for _, rSvc := range cm.containers {
		err := cm.dockerClient.ContainerKill(context.Background(), rSvc.ContainerID, "SIGKILL")
		if err != nil {
			errors = append(errors, err)
		}
		err = cm.dockerClient.ContainerRemove(context.Background(), rSvc.ContainerID, types.ContainerRemoveOptions{})
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (cm *DockerContainerManager) getUnusedPort() int {
	// get random port between 8000 and 9000
	// check if port is in use
	for {
		port := rand.Intn(1000) + 8000
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
		log.Debug().Msg("Waiting for container to start...")
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d", rSvc.AssignedPort))
		if err != nil {
			fmt.Println(err.Error())
		}
		if resp != nil && resp.StatusCode == 200 {
			log.Debug().Msg("container ready...")
			log.Info().Int64("duration_ms", time.Since(start).Milliseconds()).Msg("Container started\n")
			rSvc.Ready = true
			return true
		}
		log.Debug().Msg("Container not ready yet...")
		time.Sleep(1 * time.Second)
	}
	return false
}
