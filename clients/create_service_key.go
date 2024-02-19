package clients

import (
	"bytes"
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

// CreateServiceKey create Cloud Foundry service key
func CreateServiceKey(cliConnection plugin.CliConnection, serviceInstanceGUID string, parameters interface{}) (*models.CFServiceKey, error) {
	var apiEndpoint string
	var accessToken string
	var request *http.Request
	var response *http.Response
	var serviceKey models.CFServiceKey
	var serviceKeyCredentials models.CFCredentials
	var err error
	var url string
	var serviceParameters string
	var body []byte
	var job models.CFJob
	var link models.CFLink
	var ok bool

	t := strconv.FormatInt(time.Now().Unix(), 10)
	apiEndpoint, err = cliConnection.ApiEndpoint()
	if err != nil {
		return nil, err
	}
	accessToken, err = cliConnection.AccessToken()
	if err != nil {
		return nil, err
	}
	url = apiEndpoint + "/v3/service_credential_bindings"
	if parameters != nil {
		parametersBytes, err := json.Marshal(parameters)
		if err != nil {
			return nil, err
		}
		serviceParameters = "\"parameters\":" + string(parametersBytes) + ","
	} else {
		serviceParameters = ""
	}
	body = []byte("{" + serviceParameters + "\"type\":\"key\",\"name\":\"" + "html5-key-" + t + "\",\"relationships\":{\"service_instance\":{\"data\":{\"guid\":\"" + serviceInstanceGUID + "\"}}}}")

	log.Tracef("Making request to: %s %s\n", url, string(body))
	request, err = http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	client, err := GetDefaultClient()
	if err != nil {
		return nil, err
	}
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err = io.ReadAll(response.Body)
	log.Trace(log.Response{Head: response, Body: body})
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 202 {
		return nil, fmt.Errorf("Could not create service key: [%d] %s", response.StatusCode, string(body))
	}

	// Pool job
	job, err = PollJob(cliConnection, response.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	// Get link to service key from job
	if link, ok = job.Links["service_credential_binding"]; !ok {
		return nil, fmt.Errorf("Malformed job resource. No 'service_credential_binding' link: %+v", job)
	}

	// Get service key
	serviceKey, err = GetServiceKeyByUrl(cliConnection, *link.Href)
	if err != nil {
		return nil, err
	}

	// Get service key details
	serviceKeyCredentials, err = GetServiceKeyDetails(cliConnection, serviceKey.GUID)
	if err != nil {
		return nil, err
	}

	// Enrich service key with credentials from details
	serviceKey.Credentials = serviceKeyCredentials

	return &serviceKey, nil
}
