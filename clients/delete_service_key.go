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
	var errorResponseObject models.CFErrorResponse
	var currentTry = 1

	url = "/v2/service_keys/" + serviceKeyGUID

	for currentTry <= maxRetryCount {
		log.Tracef("Making request to (try %d/%d): %s\n", currentTry, maxRetryCount, url)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "DELETE")
		if err != nil {
			return err
		}

		responseBytes = []byte(strings.Join(responseStrings, ""))
		if len(responseBytes) > 0 {
			log.Tracef("Response is not empty, maybe error: %+v\n", responseStrings)
			err = json.Unmarshal(responseBytes, &errorResponseObject)
			if err != nil {
				return err
			}
			if errorResponseObject.ErrorCode != nil {
				if currentTry == maxRetryCount {
					if errorResponseObject.Description != nil {
						return errors.New(*errorResponseObject.Description)
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
