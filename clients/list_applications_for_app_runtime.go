package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"io"
	"net/http"
)

// ListApplicationsForAppRuntime list HTML5 applications for app-runtime
func ListApplicationsForAppRuntime(serviceURL string, accessToken string) (models.HTML5ListApplicationsResponse, error) {
	var html5Response models.HTML5ListApplicationsResponse
	var request *http.Request
	var response *http.Response
	var err error
	var html5URL string
	var body []byte

	html5URL = serviceURL + "/applications/metadata/"

	log.Tracef("Making request to: %s\n", html5URL)

	client, err := GetDefaultClient()
	if err != nil {
		return html5Response, err
	}
	request, err = http.NewRequest("GET", html5URL, nil)
	if err != nil {
		return html5Response, err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	response, err = client.Do(request)
	if err != nil {
		return html5Response, err
	}

	// Get response body
	defer response.Body.Close()
	body, err = io.ReadAll(response.Body)
	log.Trace(log.Response{Head: response, Body: body})
	if err != nil {
		return html5Response, err
	}

	// Parse response JSON
	err = json.Unmarshal(body, &html5Response)
	if err != nil {
		return html5Response, err
	}

	return html5Response, nil
}
