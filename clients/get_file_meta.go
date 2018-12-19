package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"net/http"
	"strconv"
	"time"
)

// GetFileMeta get file size and etag
func GetFileMeta(serviceURL string, filePath string, accessToken string, appHostGUID string) (models.HTML5ApplicationFileMetadata, error) {
	var metaData models.HTML5ApplicationFileMetadata
	var request *http.Request
	var response *http.Response
	var err error
	var html5URL string

	html5URL = serviceURL + filePath

	log.Tracef("Making HEAD request to: %s", html5URL)

	client := &http.Client{}
	request, err = http.NewRequest("HEAD", html5URL, nil)
	if err != nil {
		return metaData, err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	if appHostGUID != "" {
		request.Header.Add("x-app-host-id", appHostGUID)
	}
	startTime := time.Now()
	response, err = client.Do(request)
	elapsedTime := time.Since(startTime)
	log.Tracef(" (%v)\n", elapsedTime)
	if err != nil {
		return metaData, err
	}

	metaData.ETag = response.Header.Get("Etag")
	metaData.FileSize, err = strconv.Atoi(response.Header.Get("Content-Length"))
	if err != nil {
		return metaData, err
	}

	return metaData, nil
}
