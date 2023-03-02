package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetEnvironment get Cloud Foundry application environment
func GetEnvironment(cliConnection plugin.CliConnection, appGUID string) (*models.CFEnvironmentResponse, error) {
	var responseObject models.CFEnvironmentResponse
	var responseStrings []string
	var err error
	var url string

	url = "/v3/apps/" + appGUID + "/env"

	log.Tracef("Making request to: %s\n", url)
	responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(strings.Join(responseStrings, "")), &responseObject)
	if err != nil {
		return nil, err
	}

	return &responseObject, nil
}
