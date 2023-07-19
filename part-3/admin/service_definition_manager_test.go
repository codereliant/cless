package admin

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRegisterServiceDefinition tests the RegisterServiceDefinition method
func TestRegisterServiceDefinition(t *testing.T) {
	repo := NewInMemoryServiceDefinitionRepository()
	serviceDefinitionManager := NewServiceDefinitionManager(repo)
	err := serviceDefinitionManager.RegisterServiceDefinition("test", "test", "test", 8080, "")
	if err != nil {
		t.Errorf("Failed to register service definition: %s", err)
	}
	// get the service definition
	serviceDefinition, err := repo.GetByName("test")
	if err != nil {
		t.Errorf("Failed to get service definition: %s", err)
	}
	assert.Equal(t, "test", serviceDefinition.Name)
	assert.Equal(t, "test", serviceDefinition.ImageName)
	assert.Equal(t, "test", serviceDefinition.ImageTag)
	assert.Equal(t, 8080, serviceDefinition.Port)
	assert.Regexp(t, regexp.MustCompile(`app-[0-9]+\.cless\.cloud`), serviceDefinition.Host)
}

// TestRegisterServiceDefinitionWithHost tests the RegisterServiceDefinition method with a host
func TestRegisterServiceDefinitionWithHost(t *testing.T) {
	repo := NewInMemoryServiceDefinitionRepository()
	serviceDefinitionManager := NewServiceDefinitionManager(repo)
	err := serviceDefinitionManager.RegisterServiceDefinition("test", "test", "test", 8080, "test.cless.cloud")
	if err != nil {
		t.Errorf("Failed to register service definition: %s", err)
	}

	// get the service definition
	serviceDefinition, err := repo.GetByName("test")
	if err != nil {
		t.Errorf("Failed to get service definition: %s", err)
	}
	assert.Equal(t, "test.cless.cloud", serviceDefinition.Host)
}

// TestListServiceDefinitions tests the ListServiceDefinitions method
func TestListServiceDefinitions(t *testing.T) {
	repo := NewInMemoryServiceDefinitionRepository()
	serviceDefinitionManager := NewServiceDefinitionManager(repo)
	err := serviceDefinitionManager.RegisterServiceDefinition("test", "test", "test", 8080, "")
	if err != nil {
		t.Errorf("Failed to register service definition: %s", err)
	}
	err = serviceDefinitionManager.RegisterServiceDefinition("test2", "test", "test", 8080, "")
	if err != nil {
		t.Errorf("Failed to register service definition: %s", err)
	}
	serviceDefinitions, err := serviceDefinitionManager.ListAllServiceDefinitions()
	if err != nil {
		t.Errorf("Failed to list service definitions: %s", err)
	}
	assert.Equal(t, 2, len(serviceDefinitions))
}
