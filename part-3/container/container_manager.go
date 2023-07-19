package container

import (
	"fmt"
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
	StopAndRemoveAllContainers() []error
}
