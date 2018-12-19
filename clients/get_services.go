package clients

import (
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

	services = make([]models.CFService, 0)
	firstURL := "/v2/services"
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
			services = append(services, models.CFService{Name: *service.Entity.Label, GUID: service.Metadata.GUID})
		}
		nextURL = responseObject.NextURL
	}

	return services, nil
}
