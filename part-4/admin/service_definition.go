package admin

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceAlreadyExists = errors.New("service already exists")
var randLock = &sync.Mutex{}

type ServiceDefinition struct {
	gorm.Model
	Name           string           `json:"name" gorm:"unique"`
	Versions       []ServiceVersion `json:"versions" gorm:"foreignKey:ServiceDefinitionID"`
	TrafficWeights []TrafficWeight  `json:"traffic_weights" gorm:"foreignKey:ServiceDefinitionID"`
	Host           string           `json:"host"`
}

type TrafficWeight struct {
	gorm.Model
	ServiceDefinitionID uint                        `json:"service_definition_id" gorm:"index,references:ID"`
	Weights             datatypes.JSONSlice[Weight] `json:"weights"`
}

type Weight struct {
	ServiceVersionID uint `json:"service_version_id"`
	Weight           uint `json:"weight"`
}

type ServiceVersion struct {
	gorm.Model
	ServiceDefinitionID uint                        `json:"service_definition_id" gorm:"index,references:ID"`
	ImageName           string                      `json:"image_name"`
	ImageTag            string                      `json:"image_tag"`
	Port                int                         `json:"port"`
	EnvVars             datatypes.JSONSlice[string] `json:"env_vars"`
}

type ExternalServiceDefinition struct {
	Sdef    *ServiceDefinition
	Version *ServiceVersion
}

// method on ExternalServiceDefinition to create a unique key of the form <service_id>:<version_id>
func (sDef *ExternalServiceDefinition) GetKey() string {
	return fmt.Sprintf("%d:%d", sDef.Sdef.ID, sDef.Version.ID)
}

func (sDef *ServiceDefinition) isValid() bool {
	return (sDef.Name != "" && sDef.Name != "admin")
}

func (sVer *ServiceVersion) isValid() bool {
	return sVer.ImageName != "" && sVer.ImageTag != "" && sVer.Port > 0
}

func (tw *TrafficWeight) isValid() bool {
	// sum of weights should be 100
	sum := 0
	for _, w := range tw.Weights {
		sum += int(w.Weight)
	}
	return sum == 100
}

// ChooseVersion randomly chooses a version based on the weights
// weight have the form of a slice of {version, weight} pairs
// sum of weights is always 100
// example: [{1, 50}, {2, 50}] means 50% of the traffic goes to version 1 and 50% to version 2
// example: [{1, 10}, {2, 20}, {3, 70}] means 10% of the traffic goes to version 1, 20% to version 2 and 70% to version 3
// example: [{1, 100}] means 100% of the traffic goes to version 1
// example: [{1, 50}, {2, 50}, {3, 50}] is invalid because the sum of weights is 150
// method needs to be concurrency safe
func (sDef *ServiceDefinition) ChooseVersion() uint {
	randLock.Lock()
	r := rand.Intn(101)
	randLock.Unlock()

	// grab latest of traffic weights
	tw := sDef.TrafficWeights[len(sDef.TrafficWeights)-1]
	for _, w := range tw.Weights {
		r -= int(w.Weight)
		if r <= 0 {
			return w.ServiceVersionID
		}
	}
	return 0
}

type ServiceDefinitionRepository interface {
	GetAll() ([]ServiceDefinition, error)
	GetByName(name string) (*ServiceDefinition, error)
	GetByHostName(hostName string) (*ServiceDefinition, error)
	Create(service ServiceDefinition) error
	AddVersion(service *ServiceDefinition, version *ServiceVersion) error
	AddTrafficWeight(service *ServiceDefinition, weight *TrafficWeight) error
}
