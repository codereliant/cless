package admin

import (
	"gorm.io/gorm"
)

// create type for the sqlite repository that will implement the ServiceDefinitionRepository interface
type SqliteServiceDefinitionRepository struct {
	db *gorm.DB
}

// create a constructor for the sqlite repository
func NewSqliteServiceDefinitionRepository(db *gorm.DB) ServiceDefinitionRepository {
	db.AutoMigrate(&ServiceDefinition{})
	db.AutoMigrate(&ServiceVersion{})
	db.AutoMigrate(&TrafficWeight{})
	return &SqliteServiceDefinitionRepository{db: db}
}

// implement the GetAll method
func (r *SqliteServiceDefinitionRepository) GetAll() ([]ServiceDefinition, error) {
	var services []ServiceDefinition
	result := r.db.Preload("Versions").Preload("TrafficWeights").Find(&services)
	if result.Error != nil {
		return nil, result.Error
	}
	return services, nil
}

// implement the GetByName method
func (r *SqliteServiceDefinitionRepository) GetByName(name string) (*ServiceDefinition, error) {
	var service ServiceDefinition
	result := r.db.Preload("Versions").Preload("TrafficWeights").First(&service, "name = ?", name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// implement the GetByHostName method
func (r *SqliteServiceDefinitionRepository) GetByHostName(hostName string) (*ServiceDefinition, error) {
	var service ServiceDefinition
	result := r.db.Preload("Versions").Preload("TrafficWeights").First(&service, "host = ?", hostName)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// implement the Create method
func (r *SqliteServiceDefinitionRepository) Create(service ServiceDefinition) error {
	result := r.db.Create(&service)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// AddVersion create new version and add it to the service
func (r *SqliteServiceDefinitionRepository) AddVersion(service *ServiceDefinition, version *ServiceVersion) error {
	err := r.db.Model(service).Association("Versions").Append(version)
	if err != nil {
		return err
	}
	return nil
}

// AddTrafficWeight create new traffic weight and add it to the service
func (r *SqliteServiceDefinitionRepository) AddTrafficWeight(service *ServiceDefinition, weight *TrafficWeight) error {
	err := r.db.Model(service).Association("TrafficWeights").Append(weight)
	if err != nil {
		return err
	}
	return nil
}
