package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServiceInstanceByName get Cloud Foundry service instance by name
func GetServiceInstanceByName(cliConnection plugin.CliConnection, spaceGUID string, serviceInstanceName string) (models.CFServiceInstance, error) {
	var serviceInstances []models.CFServiceInstance
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string

	serviceInstances = make([]models.CFServiceInstance, 0)
	firstURL := "/v2/service_instances?q=name:" + serviceInstanceName + "&q=space_guid:" + spaceGUID
	nextURL = &firstURL

	for nextURL != nil {
		log.Tracef("Making request to: %s\n", *nextURL)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", *nextURL)
		if err != nil {
			return models.CFServiceInstance{}, err
		}

		err = json.Unmarshal([]byte(strings.Join(responseStrings, "")), &responseObject)
		if err != nil {
			return models.CFServiceInstance{}, err
		}

		for _, serviceInstance := range responseObject.Resources {
			serviceInstances = append(serviceInstances, models.CFServiceInstance{Name: *serviceInstance.Entity.Name, GUID: serviceInstance.Metadata.GUID, UpdatedAt: serviceInstance.Metadata.UpdatedAt})
		}
		nextURL = responseObject.NextURL
	}

	if len(serviceInstances) == 0 {
		return models.CFServiceInstance{}, fmt.Errorf("Service instance with name '%s' not found", serviceInstanceName)
	}

	return serviceInstances[0], nil
}
