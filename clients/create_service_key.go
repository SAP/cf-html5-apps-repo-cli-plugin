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

// CreateServiceKey create Cloud Foundry service key
func CreateServiceKey(cliConnection plugin.CliConnection, serviceInstanceGUID string) (*models.CFServiceKey, error) {
	var serviceKey *models.CFServiceKey
	var responseObject models.CFResource
	var errorResponseObject models.CFErrorResponse
	var responseStrings []string
	var responseBytes []byte
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
	serviceKey = &models.CFServiceKey{GUID: responseObject.Metadata.GUID, Name: *responseObject.Entity.Name, Credentials: *responseObject.Entity.Credentials}

	return serviceKey, nil
}
