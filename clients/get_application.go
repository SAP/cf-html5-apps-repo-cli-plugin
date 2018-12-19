package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetApplication get Cloud Foundry application
func GetApplication(cliConnection plugin.CliConnection, spaceGUID string, appName string) (*models.CFApplication, error) {
	var application *models.CFApplication
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var url string

	url = "/v2/spaces/" + spaceGUID + "/apps?q=name%3A" + appName

	log.Tracef("Making request to: %s\n", url)
	responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(strings.Join(responseStrings, "")), &responseObject)
	if err != nil {
		return nil, err
	}
	application = &models.CFApplication{GUID: responseObject.Resources[0].Metadata.GUID, Name: *responseObject.Resources[0].Entity.Name}

	return application, nil
}
