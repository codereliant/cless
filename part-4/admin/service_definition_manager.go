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

func SetOfAvailableHosts() map[string]bool {
	hosts := make(map[string]bool)
	for i := 0; i <= 100; i++ {
		hosts[fmt.Sprintf(HostNameTemplate, i)] = true
	}
	return hosts
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

func (m *ServiceDefinitionManager) RegisterServiceDefinition(
	name string,
	host string,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	service := ServiceDefinition{
		Name: name,
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

// AddVersion adds a new version to a service definition
func (m *ServiceDefinitionManager) AddVersion(
	service *ServiceDefinition,
	version *ServiceVersion,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if !version.isValid() {
		return errors.New("invalid service version")
	}
	err := m.repo.AddVersion(service, version)
	if err != nil {
		return err
	}
	return nil
}

// AddTrafficWeight adds a new traffic weight to a service definition
func (m *ServiceDefinitionManager) AddTrafficWeight(
	service *ServiceDefinition,
	weight *TrafficWeight,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if !weight.isValid() {
		return errors.New("invalid traffic weight")
	}
	err := m.repo.AddTrafficWeight(service, weight)
	if err != nil {
		return err
	}
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

func (m *ServiceDefinitionManager) GetExternalServiceDefinitionByHost(hostname string, version uint) (*ExternalServiceDefinition, error) {
	sDef, err := m.repo.GetByHostName(hostname)
	if err != nil {
		return nil, err
	}

	var sVersion *ServiceVersion

	for _, v := range sDef.Versions {
		if v.ID == version {
			sVersion = &v
			break
		}
	}

	if sVersion == nil {
		return nil, fmt.Errorf("version %d not found for service %s and hostname %s", version, sDef.Name, hostname)
	}

	return &ExternalServiceDefinition{
		Sdef:    sDef,
		Version: sVersion,
	}, nil
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
