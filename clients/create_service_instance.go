package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

// CreateServiceInstance create Cloud Foundry service instance
func CreateServiceInstance(cliConnection plugin.CliConnection, spaceGUID string, servicePlan models.CFServicePlan, parameters interface{}) (*models.CFServiceInstance, error) {
	var serviceInstance *models.CFServiceInstance
	var responseObject models.CFResource
	var errorResponseObject models.CFErrorResponse
	var responseStrings []string
	var responseBytes []byte
	var err error
	var url string
	var serviceParameters string
	var body string

	t := strconv.FormatInt(time.Now().Unix(), 10)
	url = "/v2/service_instances"
	if parameters != nil {
		parametersBytes, err := json.Marshal(parameters)
		if err != nil {
			return nil, err
		}
		serviceParameters = "\"parameters\":" + string(parametersBytes) + ","
	} else {
		serviceParameters = ""
	}
	body = "'{" + serviceParameters + "\"space_guid\":\"" + spaceGUID + "\",\"name\":\"" + servicePlan.Name + "-" + t + "\",\"service_plan_guid\":\"" + servicePlan.GUID + "\"}'"

	log.Tracef("Making request to: %s\n", url)
	responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "POST", "-d", body)
	if err != nil {
		return nil, err
	}

	responseBytes = []byte(strings.Join(responseStrings, ""))
	err = json.Unmarshal(responseBytes, &responseObject)
	if err != nil {
		return nil, err
	}
	if responseObject.Entity == nil {
		err = json.Unmarshal(responseBytes, &errorResponseObject)
		if err != nil {
			return nil, err
		}
		if errorResponseObject.Description != nil {
			return nil, errors.New("\n" + *errorResponseObject.Description)
		}
		return nil, errors.New(strings.Join(responseStrings, "\n"))
	}
	serviceInstance = &models.CFServiceInstance{GUID: responseObject.Metadata.GUID, Name: *responseObject.Entity.Name}

	return serviceInstance, nil
}
