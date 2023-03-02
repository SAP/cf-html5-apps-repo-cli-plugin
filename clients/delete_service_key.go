package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// DeleteServiceKey delete Cloud Foundry service key
func DeleteServiceKey(cliConnection plugin.CliConnection, serviceKeyGUID string, maxRetryCount int) error {
	var err error
	var url string
	var responseStrings []string
	var responseBytes []byte
	var errorResponse models.CFErrorResponse
	var currentTry = 1

	url = "/v3/service_credential_bindings/" + serviceKeyGUID

	for currentTry <= maxRetryCount {
		log.Tracef("Making request to (try %d/%d): %s\n", currentTry, maxRetryCount, url)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "DELETE")
		if err != nil {
			return err
		}

		responseBytes = []byte(strings.Join(responseStrings, ""))
		if len(responseBytes) > 0 {
			log.Tracef("Response is not empty, maybe error: %+v\n", responseStrings)
			err = json.Unmarshal(responseBytes, &errorResponse)
			if err != nil {
				return err
			}
			if len(errorResponse) == 0 {
				return errors.New(strings.Join(responseStrings, "\n"))
			}
			if errorResponse[0].Code > 0 {
				if currentTry == maxRetryCount {
					if errorResponse[0].Title != "" || errorResponse[0].Detail != "" {
						return errors.New(errorResponse[0].Title + " " + errorResponse[0].Detail)
					}
					return errors.New(strings.Join(responseStrings, "\n"))
				}
				currentTry++
				continue
			}
		} else {
			return nil
		}
	}

	return err
}
