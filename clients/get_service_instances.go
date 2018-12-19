package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServiceInstances get Cloud Foundry service instances
func GetServiceInstances(cliConnection plugin.CliConnection, spaceGUID string, servicePlans []models.CFServicePlan) ([]models.CFServiceInstance, error) {
	var serviceInstances []models.CFServiceInstance
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string
	var servicePlanGUIDs []string

	serviceInstances = make([]models.CFServiceInstance, 0)
	servicePlanGUIDs = make([]string, 0)
	for _, servicePlan := range servicePlans {
		servicePlanGUIDs = append(servicePlanGUIDs, servicePlan.GUID)
	}
	firstURL := "/v2/service_instances?q=service_plan_guid%20IN%20" + strings.Join(servicePlanGUIDs, ",") + "%3Bspace_guid:" + spaceGUID
	nextURL = &firstURL

	for nextURL != nil {
		log.Tracef("Making request to: %s\n", *nextURL)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", *nextURL)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(strings.Join(responseStrings, "")), &responseObject)
		if err != nil {
			return nil, err
		}

		for _, serviceInstance := range responseObject.Resources {
			serviceInstances = append(serviceInstances, models.CFServiceInstance{Name: *serviceInstance.Entity.Name, GUID: serviceInstance.Metadata.GUID, UpdatedAt: serviceInstance.Metadata.UpdatedAt})
		}
		nextURL = responseObject.NextURL
	}

	return serviceInstances, nil
}
