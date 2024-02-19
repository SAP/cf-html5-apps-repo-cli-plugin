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
	var pathStart int
	var pathSlice string
	var servicePlanGUIDs []string

	serviceInstances = make([]models.CFServiceInstance, 0)
	servicePlanGUIDs = make([]string, 0)
	for _, servicePlan := range servicePlans {
		servicePlanGUIDs = append(servicePlanGUIDs, servicePlan.GUID)
	}
	firstURL := "/v3/service_instances?service_plan_guids=" + strings.Join(servicePlanGUIDs, ",") + "&space_guids=" + spaceGUID
	nextURL = &firstURL

	for nextURL != nil {
		log.Tracef("Making request to: %s\n", *nextURL)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", *nextURL)
		if err != nil {
			return nil, err
		}

		responseObject = models.CFResponse{}
		body := []byte(strings.Join(responseStrings, ""))
		log.Trace(log.Response{Body: body})
		err = json.Unmarshal(body, &responseObject)
		if err != nil {
			return nil, err
		}

		for _, serviceInstance := range responseObject.Resources {
			serviceInstances = append(serviceInstances, models.CFServiceInstance{
				Name:          serviceInstance.Name,
				GUID:          serviceInstance.GUID,
				UpdatedAt:     serviceInstance.UpdatedAt,
				LastOperation: serviceInstance.LastOperation,
			})
		}
		if responseObject.Pagination.Next.Href != nil && *nextURL == *responseObject.Pagination.Next.Href {
			log.Tracef("Unexpected value of the next page URL (equal to previous): %s\n", *nextURL)
			break
		}
		nextURL = responseObject.Pagination.Next.Href
		if nextURL != nil {
			pathStart = strings.Index(*nextURL, "/v3/service_instances")
			if pathStart > 0 {
				pathSlice = (*nextURL)[pathStart:]
				nextURL = &pathSlice
			}
		}
	}

	return serviceInstances, nil
}
