package admin

import (
	"errors"
	"fmt"
	"sync"
)

const HostNameTemplate = "app-%d.cless.cloud"

type ServiceDefinitionManager struct {
	repo  ServiceDefinitionRepository
	hosts map[string]bool
	mutex sync.Mutex
}

func NewServiceDefinitionManager(repo ServiceDefinitionRepository) *ServiceDefinitionManager {
	hosts := SetOfAvailableHosts()
	sDefs, err := repo.GetAll()
	if err != nil {
		panic(err)
	}
	for _, sDef := range sDefs {
		delete(hosts, sDef.Host)
	}
	return &ServiceDefinitionManager{
		repo:  repo,
		hosts: hosts,
		mutex: sync.Mutex{},
	}
}

func (m *ServiceDefinitionManager) RegisterServiceDefinition(name string, imageName string, imageTag string, port int, host string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	service := ServiceDefinition{
		Name:      name,
		ImageName: imageName,
		ImageTag:  imageTag,
		Port:      port,
	}
	if host != "" {
		service.Host = host
	} else {
		h, err := m.NewHostName()
		if err != nil {
			return err
		}
		service.Host = *h
	}
	err := m.repo.Create(service)
	if err != nil {
		return err
	}

	delete(m.hosts, service.Host)
	return nil
}

func (m *ServiceDefinitionManager) ListAllServiceDefinitions() ([]ServiceDefinition, error) {
	return m.repo.GetAll()
}

func (m *ServiceDefinitionManager) GetServiceDefinitionByName(name string) (*ServiceDefinition, error) {
	return m.repo.GetByName(name)
}

func (m *ServiceDefinitionManager) GetServiceDefinitionByHost(hostname string) (*ServiceDefinition, error) {
	return m.repo.GetByHostName(hostname)
}

func (m *ServiceDefinitionManager) NewHostName() (*string, error) {
	if len(m.hosts) == 0 {
		return nil, errors.New("no more hosts available")
	}
	for host := range m.hosts {
		return &host, nil
	}
	return nil, errors.New("no more hosts available")
}

func SetOfAvailableHosts() map[string]bool {
	hosts := make(map[string]bool)
	for i := 0; i <= 100; i++ {
		hosts[fmt.Sprintf(HostNameTemplate, i)] = true
	}
	return hosts
}
