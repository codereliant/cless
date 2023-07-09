package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

const AdminHost = "admin.cless.cloud"
const AdminPort = 1323
const HostNameTemplate = "%s.cless.cloud"

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceAlreadyExists = errors.New("service already exists")

type ServiceDefinition struct {
	Name      string `json:"name"`
	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`
	Port      int    `json:"port"`
	Host      string `json:"host"`
}

func (sDef *ServiceDefinition) isValid() bool {
	return (sDef.Name != "" && sDef.Name != "admin") && sDef.ImageName != "" && sDef.ImageTag != "" && sDef.Port > 0
}

type ServiceDefinitionRepository interface {
	GetAll() ([]ServiceDefinition, error)
	GetByName(name string) (*ServiceDefinition, error)
	Create(service ServiceDefinition) error
	Update(service ServiceDefinition) error
}

type InMemoryServiceDefinitionRepository struct {
	services map[string]ServiceDefinition
	mutex    *sync.Mutex
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

type ServiceDefinitionManager struct {
	repo ServiceDefinitionRepository
}

func NewServiceDefinitionManager(repo ServiceDefinitionRepository) *ServiceDefinitionManager {
	return &ServiceDefinitionManager{
		repo: repo,
	}
}

func (m *ServiceDefinitionManager) RegisterServiceDefinition(name string, imageName string, imageTag string, port int) error {
	service := ServiceDefinition{
		Name:      name,
		ImageName: imageName,
		ImageTag:  imageTag,
		Port:      port,
	}
	service.Host = fmt.Sprintf(HostNameTemplate, name)
	return m.repo.Create(service)
}

func (m *ServiceDefinitionManager) ListAllServiceDefinitions() ([]ServiceDefinition, error) {
	return m.repo.GetAll()
}

func (m *ServiceDefinitionManager) GetServiceDefinitionByName(name string) (*ServiceDefinition, error) {
	return m.repo.GetByName(name)
}

func StartAdminServer(manager *ServiceDefinitionManager) {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Admin server is running")
	})

	e.GET("/serviceDefinitions", func(c echo.Context) error {
		services, err := manager.ListAllServiceDefinitions()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, services)
	})

	e.GET("/serviceDefinitions/:name", func(c echo.Context) error {
		name := c.Param("name")
		service, err := manager.GetServiceDefinitionByName(name)
		if err == ErrServiceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		return c.JSON(http.StatusOK, service)
	})

	e.POST("/serviceDefinitions", func(c echo.Context) error {
		service := new(ServiceDefinition)
		if err := c.Bind(service); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		if !service.isValid() {
			return c.String(http.StatusBadRequest, "Invalid service definition")
		}
		if err := manager.RegisterServiceDefinition(service.Name, service.ImageName, service.ImageTag, service.Port); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusCreated, "Service definition created")
	})
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", AdminPort)))
}
