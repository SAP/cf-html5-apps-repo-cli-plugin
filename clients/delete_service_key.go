package clients

import (
	"bytes"
	"cf-html5-apps-repo-cli-plugin/log"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudfoundry/cli/plugin"
)

// DeleteServiceKey delete Cloud Foundry service key
func DeleteServiceKey(cliConnection plugin.CliConnection, serviceKeyGUID string, maxRetryCount int) error {
	var err error
	var url string
	var currentTry = 1
	var request *http.Request
	var response *http.Response
	var body []byte
	var apiEndpoint string
	var accessToken string

	apiEndpoint, err = cliConnection.ApiEndpoint()
	if err != nil {
		return err
	}

	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return err
	}

	url = apiEndpoint + "/v3/service_credential_bindings/" + serviceKeyGUID

	for currentTry <= maxRetryCount {
		log.Tracef("Making request to: %s (try %d/%d)\n", url, currentTry, maxRetryCount)
		request, err = http.NewRequest("DELETE", url, bytes.NewBuffer([]byte{}))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", accessToken)

		client, err := GetDefaultClient()
		if err != nil {
			return err
		}
		response, err = client.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		body, err = io.ReadAll(response.Body)
		log.Trace(log.Response{Head: response, Body: body})
		if err != nil {
			return err
		}

		if response.StatusCode != 202 {
			if currentTry != maxRetryCount {
				continue
			}
			return fmt.Errorf("Could not delete service key: [%d] %s", response.StatusCode, string(body[:]))
		} else {
			// Pool job
			_, err = PollJob(cliConnection, response.Header.Get("Location"))
			return err
		}
	}

	return err
}
