package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServiceInstancesByNamePrefix get Cloud Foundry service instance by name
func GetServiceInstancesByNamePrefix(cliConnection plugin.CliConnection, spaceGUID string, serviceInstancesNamePrefix string) ([]models.CFServiceInstance, error) {
	var serviceInstances []models.CFServiceInstance
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string

	serviceInstances = make([]models.CFServiceInstance, 0)
	firstURL := "/v3/service_instances?space_guids=" + spaceGUID
	nextURL = &firstURL

	// Remove placeholder
	if serviceInstancesNamePrefix[len(serviceInstancesNamePrefix)-1:] == "*" {
		serviceInstancesNamePrefix = serviceInstancesNamePrefix[:len(serviceInstancesNamePrefix)-1]
	}

	for nextURL != nil {
		log.Tracef("Making request to: %s\n", *nextURL)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", *nextURL)
		if err != nil {
			return serviceInstances, err
		}

		responseObject = models.CFResponse{}
		body := []byte(strings.Join(responseStrings, ""))
		log.Trace(log.Response{Body: body})
		err = json.Unmarshal(body, &responseObject)
		if err != nil {
			return serviceInstances, err
		}

		for _, serviceInstance := range responseObject.Resources {
			name := serviceInstance.Name
			if len(name) >= len(serviceInstancesNamePrefix) && name[0:len(serviceInstancesNamePrefix)] == serviceInstancesNamePrefix {
				serviceInstances = append(serviceInstances, models.CFServiceInstance{
					Name:          serviceInstance.Name,
					GUID:          serviceInstance.GUID,
					UpdatedAt:     serviceInstance.UpdatedAt,
					LastOperation: serviceInstance.LastOperation,
				})
			}
		}
		if responseObject.Pagination.Next.Href != nil && *nextURL == *responseObject.Pagination.Next.Href {
			log.Tracef("Unexpected value of the next page URL (equal to previous): %s\n", *nextURL)
			break
		}
		nextURL = responseObject.Pagination.Next.Href
	}

	if len(serviceInstances) == 0 {
		return serviceInstances, fmt.Errorf("Service instances with name prefix '%s' not found", serviceInstancesNamePrefix)
	}

	return serviceInstances, nil
}
