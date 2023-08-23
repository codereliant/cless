package container

import (
	"fmt"
	"time"
)

type RunningService struct {
	ContainerID      string    // docker container ID
	AssignedPort     int       // port assigned to the container
	Ready            bool      // whether the container is ready to serve requests
	LastTimeAccessed time.Time // last time the container was accessed
}

func (rSvc *RunningService) GetHost() string {
	return fmt.Sprintf("localhost:%d", rSvc.AssignedPort)
}

type ContainerManager interface {
	GetRunningServiceForHost(host string, version uint) (*string, error)
	StopAndRemoveAllContainers() []error
}
