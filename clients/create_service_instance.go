package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

// CreateServiceInstance create Cloud Foundry service instance
func CreateServiceInstance(cliConnection plugin.CliConnection, spaceGUID string, servicePlan models.CFServicePlan) (*models.CFServiceInstance, error) {
	var serviceInstance *models.CFServiceInstance
	var responseObject models.CFResource
	var responseStrings []string
	var err error
	var url string
	var body string

	t := strconv.FormatInt(time.Now().Unix(), 10)
	url = "/v2/service_instances"
	body = "'{\"space_guid\":\"" + spaceGUID + "\",\"name\":\"" + servicePlan.Name + "-" + t + "\",\"service_plan_guid\":\"" + servicePlan.GUID + "\"}'"

	log.Tracef("Making request to: %s\n", url)
	responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "POST", "-d", body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(strings.Join(responseStrings, "")), &responseObject)
	if err != nil {
		return nil, err
	}
	serviceInstance = &models.CFServiceInstance{GUID: responseObject.Metadata.GUID, Name: *responseObject.Entity.Name}

	return serviceInstance, nil
}
