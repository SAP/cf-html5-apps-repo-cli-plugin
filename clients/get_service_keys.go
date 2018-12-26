package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServiceKeys get Cloud Foundry service keys
func GetServiceKeys(cliConnection plugin.CliConnection, serviceInstanceGUID string) ([]models.CFServiceKey, error) {
	var serviceKeys []models.CFServiceKey
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string

	serviceKeys = make([]models.CFServiceKey, 0)
	firstURL := "/v2/service_instances/" + serviceInstanceGUID + "/service_keys"
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

		for _, serviceKey := range responseObject.Resources {
			serviceKeys = append(serviceKeys, models.CFServiceKey{Name: *serviceKey.Entity.Name, GUID: serviceKey.Metadata.GUID, Credentials: *serviceKey.Entity.Credentials})
		}
		nextURL = responseObject.NextURL
	}

	return serviceKeys, nil
}
