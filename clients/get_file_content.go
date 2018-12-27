package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"io/ioutil"
	"net/http"
)

// GetFileContent get HTML5 applications file content
func GetFileContent(serviceURL string, filePath string, accessToken string, appHostGUID string, resultChannel chan<- models.HTML5ApplicationFileContent) {
	var request *http.Request
	var response *http.Response
	var err error
	var html5URL string
	var body = make([]byte, 0)

	html5URL = serviceURL + filePath

	log.Tracef("Making request to: %s\n", html5URL)

	client := &http.Client{}
	request, err = http.NewRequest("GET", html5URL, nil)
	if err != nil {
		resultChannel <- models.HTML5ApplicationFileContent{Error: err}
		return
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	if appHostGUID != "" {
		request.Header.Add("x-app-host-id", appHostGUID)
	}
	response, err = client.Do(request)
	if err != nil {
		resultChannel <- models.HTML5ApplicationFileContent{Error: err}
		return
	}

	// Get response body
	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		resultChannel <- models.HTML5ApplicationFileContent{Error: err}
		return
	}
	resultChannel <- models.HTML5ApplicationFileContent{Content: body}
}
