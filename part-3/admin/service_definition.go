package admin

import (
	"errors"

	"gorm.io/gorm"
)

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceAlreadyExists = errors.New("service already exists")

type ServiceDefinition struct {
	gorm.Model
	Name      string `json:"name" gorm:"unique"`
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
	GetByHostName(hostName string) (*ServiceDefinition, error)
	Create(service ServiceDefinition) error
}
