package admin

import "sync"

type InMemoryServiceDefinitionRepository struct {
	services map[string]ServiceDefinition
	mutex    *sync.Mutex
}

func NewInMemoryServiceDefinitionRepository() ServiceDefinitionRepository {
	return &InMemoryServiceDefinitionRepository{
		services: make(map[string]ServiceDefinition),
		mutex:    &sync.Mutex{},
	}
}

func (r *InMemoryServiceDefinitionRepository) GetAll() ([]ServiceDefinition, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	services := make([]ServiceDefinition, 0)
	for _, service := range r.services {
		services = append(services, service)
	}
	return services, nil
}

func (r *InMemoryServiceDefinitionRepository) GetByName(name string) (*ServiceDefinition, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	service, ok := r.services[name]
	if !ok {
		return nil, ErrServiceNotFound
	}
	return &service, nil
}

func (r *InMemoryServiceDefinitionRepository) GetByHostName(hostName string) (*ServiceDefinition, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, service := range r.services {
		if service.Host == hostName {
			return &service, nil
		}
	}
	return nil, ErrServiceNotFound
}

func (r *InMemoryServiceDefinitionRepository) Create(service ServiceDefinition) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.services[service.Name]
	if ok {
		return ErrServiceAlreadyExists
	}
	r.services[service.Name] = service
	return nil
}

func (r *InMemoryServiceDefinitionRepository) Update(service ServiceDefinition) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.services[service.Name]
	if !ok {
		return ErrServiceNotFound
	}
	r.services[service.Name] = service
	return nil
}
