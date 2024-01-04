package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"net/http"
	"strconv"
	"time"
)

// GetFileMeta get file size and etag
func GetFileMeta(serviceURL string, filePath string, accessToken string, appHostGUID string, resultChannel chan<- models.HTML5ApplicationFileMetadata) {
	var metaData models.HTML5ApplicationFileMetadata
	var request *http.Request
	var response *http.Response
	var err error
	var html5URL string

	html5URL = serviceURL + filePath

	log.Tracef("Making HEAD request to: %s\n", html5URL)

	client, err := GetDefaultClient()
	if err != nil {
		metaData.Error = err
		resultChannel <- metaData
		return
	}
	request, err = http.NewRequest("HEAD", html5URL, nil)
	if err != nil {
		metaData.Error = err
		resultChannel <- metaData
		return
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	if appHostGUID != "" {
		request.Header.Add("x-app-host-id", appHostGUID)
	}
	startTime := time.Now()
	response, err = client.Do(request)
	elapsedTime := time.Since(startTime)
	log.Tracef("Request to %s took %v\n", html5URL, elapsedTime)
	if err != nil {
		metaData.Error = err
		resultChannel <- metaData
		return
	}
	defer response.Body.Close()

	metaData.ETag = response.Header.Get("Etag")
	metaData.FileSize, err = strconv.Atoi(response.Header.Get("Content-Length"))
	if err != nil {
		metaData.Error = err
		resultChannel <- metaData
		return
	}

	resultChannel <- metaData
}
