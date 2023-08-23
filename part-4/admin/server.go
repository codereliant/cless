package admin

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

const AdminHost = "admin.cless.cloud"
const AdminPort = 1323

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
		if err := manager.RegisterServiceDefinition(service.Name, service.Host); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusCreated, "Service definition created")
	})

	// add new version for a service definition
	e.POST("/serviceDefinitions/:name/versions", func(c echo.Context) error {
		name := c.Param("name")
		service, err := manager.GetServiceDefinitionByName(name)
		// bind version
		version := new(ServiceVersion)
		if err := c.Bind(version); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		if err == ErrServiceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		if err := manager.AddVersion(service, version); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusCreated, "Version added")
	})

	// list versions for a service definition
	e.GET("/serviceDefinitions/:name/versions", func(c echo.Context) error {
		name := c.Param("name")
		service, err := manager.GetServiceDefinitionByName(name)
		if err == ErrServiceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		return c.JSON(http.StatusOK, service.Versions)
	})

	// list traffic weights for a service definition
	e.GET("/serviceDefinitions/:name/trafficWeights", func(c echo.Context) error {
		name := c.Param("name")
		service, err := manager.GetServiceDefinitionByName(name)
		if err == ErrServiceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		return c.JSON(http.StatusOK, service.TrafficWeights)
	})

	// add new traffic weight for a service definition
	e.POST("/serviceDefinitions/:name/trafficWeights", func(c echo.Context) error {
		name := c.Param("name")
		service, err := manager.GetServiceDefinitionByName(name)
		// bind traffic weight
		weight := new(TrafficWeight)
		if err := c.Bind(weight); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		if err == ErrServiceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		if err := manager.AddTrafficWeight(service, weight); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusCreated, "Traffic weight added")
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", AdminPort)))
}
