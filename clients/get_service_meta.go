package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// GetServiceMeta get metadata of html5-apps-repo service app-host plan service instance
func GetServiceMeta(serviceURL string, accessToken string, resultChannel chan<- models.HTML5ServiceMeta) {
	var metaData models.HTML5ServiceMeta
	var request *http.Request
	var response *http.Response
	var body []byte
	var err error
	var html5URL string

	html5URL = serviceURL + "/app-host/metadata"

	log.Tracef("Making request to: %s\n", html5URL)

	client, err := GetDefaultClient()
	if err != nil {
		resultChannel <- models.HTML5ServiceMeta{Error: err}
		return
	}
	request, err = http.NewRequest("GET", html5URL, nil)
	if err != nil {
		resultChannel <- models.HTML5ServiceMeta{Error: err}
		return
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	startTime := time.Now()
	response, err = client.Do(request)
	elapsedTime := time.Since(startTime)
	log.Tracef("Request to %s took %v\n", html5URL, elapsedTime)
	if err != nil {
		resultChannel <- models.HTML5ServiceMeta{Error: err}
		return
	}

	// Get response body
	defer response.Body.Close()
	body, err = io.ReadAll(response.Body)
	if err != nil {
		resultChannel <- models.HTML5ServiceMeta{Error: err}
		return
	}

	// Parse response JSON
	err = json.Unmarshal(body, &metaData)
	if err != nil {
		resultChannel <- models.HTML5ServiceMeta{Error: err}
		return
	}

	resultChannel <- metaData
}
