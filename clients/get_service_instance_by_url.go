package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry/cli/plugin"
)

func GetServiceInstanceByUrl(cliConnection plugin.CliConnection, url string) (models.CFServiceInstance, error) {
	var accessToken string
	var request *http.Request
	var response *http.Response
	var err error
	var body []byte
	var serviceInstance models.CFServiceInstance

	// Setup request
	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return serviceInstance, err
	}

	log.Tracef("Making request to: %s\n", url)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return serviceInstance, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	// Make request
	client, err := GetDefaultClient()
	if err != nil {
		return serviceInstance, err
	}
	response, err = client.Do(request)
	if err != nil {
		return serviceInstance, err
	}
	defer response.Body.Close()

	// Read response body
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return serviceInstance, err
	}

	// Failed to get service instance
	if response.StatusCode != 200 {
		return serviceInstance, fmt.Errorf("Failed to get service instance by URL '%s': [%d] %+v", url, response.StatusCode, body)
	}

	// Unmarshal JSON
	err = json.Unmarshal(body, &serviceInstance)
	if err != nil {
		return serviceInstance, err
	}

	return serviceInstance, nil
}
