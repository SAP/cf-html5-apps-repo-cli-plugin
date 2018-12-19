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

// CreateServiceKey create Cloud Foundry service key
func CreateServiceKey(cliConnection plugin.CliConnection, serviceInstanceGUID string) (*models.CFServiceKey, error) {
	var serviceKey *models.CFServiceKey
	var responseObject models.CFResource
	var responseStrings []string
	var err error
	var url string
	var body string

	t := strconv.FormatInt(time.Now().Unix(), 10)
	url = "/v2/service_keys"
	body = "'{\"name\":\"html5-key-" + t + "\",\"service_instance_guid\":\"" + serviceInstanceGUID + "\"}'"

	log.Tracef("Making request to: %s\n", url)
	responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "POST", "-d", body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(strings.Join(responseStrings, "")), &responseObject)
	if err != nil {
		return nil, err
	}
	serviceKey = &models.CFServiceKey{GUID: responseObject.Metadata.GUID, Name: *responseObject.Entity.Name, Credentials: *responseObject.Entity.Credentials}

	return serviceKey, nil
}
