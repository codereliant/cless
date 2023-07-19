package admin

import "gorm.io/gorm"

// create type for the sqlite repository that will implement the ServiceDefinitionRepository interface
type SqliteServiceDefinitionRepository struct {
	db *gorm.DB
}

// create a constructor for the sqlite repository
func NewSqliteServiceDefinitionRepository(db *gorm.DB) ServiceDefinitionRepository {
	db.AutoMigrate(&ServiceDefinition{})
	return &SqliteServiceDefinitionRepository{db: db}
}

// implement the GetAll method
func (r *SqliteServiceDefinitionRepository) GetAll() ([]ServiceDefinition, error) {
	var services []ServiceDefinition
	result := r.db.Find(&services)
	if result.Error != nil {
		return nil, result.Error
	}
	return services, nil
}

// implement the GetByName method
func (r *SqliteServiceDefinitionRepository) GetByName(name string) (*ServiceDefinition, error) {
	var service ServiceDefinition
	result := r.db.First(&service, "name = ?", name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// implement the GetByHostName method
func (r *SqliteServiceDefinitionRepository) GetByHostName(hostName string) (*ServiceDefinition, error) {
	var service ServiceDefinition
	result := r.db.First(&service, "host = ?", hostName)
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
