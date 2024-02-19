package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudfoundry/cli/plugin"
)

func GetServiceKeyByUrl(cliConnection plugin.CliConnection, url string) (models.CFServiceKey, error) {
	var accessToken string
	var request *http.Request
	var response *http.Response
	var err error
	var body []byte
	var serviceKey models.CFServiceKey

	// Setup request
	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return serviceKey, err
	}

	log.Tracef("Making request to: %s\n", url)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return serviceKey, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	// Make request
	client, err := GetDefaultClient()
	if err != nil {
		return serviceKey, err
	}
	response, err = client.Do(request)
	if err != nil {
		return serviceKey, err
	}
	defer response.Body.Close()

	// Read response body
	body, err = io.ReadAll(response.Body)
	log.Trace(log.Response{Head: response, Body: body})
	if err != nil {
		return serviceKey, err
	}

	// Failed to get service key
	if response.StatusCode != 200 {
		return serviceKey, fmt.Errorf("Failed to get service key by URL '%s': [%d] %+v", url, response.StatusCode, body)
	}

	// Unmarshal JSON
	err = json.Unmarshal(body, &serviceKey)
	if err != nil {
		return serviceKey, err
	}

	return serviceKey, nil
}
