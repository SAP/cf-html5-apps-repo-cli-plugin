package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServicePlans get Cloud Foundry services
func GetServicePlans(cliConnection plugin.CliConnection, serviceGUID string) ([]models.CFServicePlan, error) {
	var servicePlans []models.CFServicePlan
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string

	servicePlans = make([]models.CFServicePlan, 0)
	firstURL := "/v2/service_plans?q=service_guid:" + serviceGUID
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

		for _, servicePlan := range responseObject.Resources {
			servicePlans = append(servicePlans, models.CFServicePlan{Name: *servicePlan.Entity.Name, GUID: servicePlan.Metadata.GUID})
		}
		nextURL = responseObject.NextURL
	}

	return servicePlans, nil
}
