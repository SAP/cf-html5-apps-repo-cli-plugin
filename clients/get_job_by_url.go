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

// GetJobByUrl get Cloud Foundry job by full URL
func GetJobByUrl(cliConnection plugin.CliConnection, url string) (models.CFJob, error) {
	var accessToken string
	var request *http.Request
	var response *http.Response
	var err error

	var body []byte
	var job models.CFJob

	// Setup request
	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return job, err
	}

	log.Tracef("Making request to: %s\n", url)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return job, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	// Make request
	client := &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		return job, err
	}
	defer response.Body.Close()

	// Read response body
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return job, err
	}

	// Failed to get job
	if response.StatusCode != 200 {
		return job, fmt.Errorf("Failed to get job by URL '%s': [%d] %+v", url, response.StatusCode, body)
	}

	// Unmarshal JSON
	err = json.Unmarshal(body, &job)
	if err != nil {
		return job, err
	}

	// Return job
	return job, nil
}
