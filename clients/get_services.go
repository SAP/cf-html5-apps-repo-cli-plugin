package clients

import (
	"cf-html5-apps-repo-cli-plugin/cache"
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServices get Cloud Foundry services
func GetServices(cliConnection plugin.CliConnection) ([]models.CFService, error) {
	var services []models.CFService
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string

	space, err := cliConnection.GetCurrentSpace()
	if err != nil {
		return nil, err
	}

	if cachedServices, ok := cache.Get("GetServices:" + space.Guid); ok {
		log.Tracef("Returning cached list of services\n")
		services = cachedServices.([]models.CFService)
		return services, nil
	}

	services = make([]models.CFService, 0)
	firstURL := "/v3/service_offerings?space_guids=" + space.Guid
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

		for _, service := range responseObject.Resources {
			services = append(services, models.CFService{
				Name: service.Name,
				GUID: service.GUID,
			})
		}
		nextURL = responseObject.Pagination.Next.Href
	}

	cache.Set("GetServices:"+space.Guid, services)

	return services, nil
}
