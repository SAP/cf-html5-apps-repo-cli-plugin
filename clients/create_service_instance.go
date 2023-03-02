package clients

import (
	"bytes"
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

// CreateServiceInstance create Cloud Foundry service instance
func CreateServiceInstance(cliConnection plugin.CliConnection, spaceGUID string, servicePlan models.CFServicePlan, parameters interface{}, name string) (*models.CFServiceInstance, error) {
	var apiEndpoint string
	var accessToken string
	var request *http.Request
	var response *http.Response
	var serviceInstance models.CFServiceInstance
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
	url = apiEndpoint + "/v3/service_instances"
	if parameters != nil {
		parametersBytes, err := json.Marshal(parameters)
		if err != nil {
			return nil, err
		}
		serviceParameters = "\"parameters\":" + string(parametersBytes) + ","
	} else {
		serviceParameters = ""
	}
	if name == "" {
		name = servicePlan.Name + "-" + t
	} else if len(name) > 1 && name[len(name)-1:] == "-" {
		name = name + servicePlan.Name + "-" + t
	}
	body = []byte("{" + serviceParameters + "\"type\":\"managed\",\"name\":\"" + name + "\",\"relationships\":{\"space\":{\"data\":{\"guid\":\"" + spaceGUID + "\"}},\"service_plan\":{\"data\":{\"guid\":\"" + servicePlan.GUID + "\"}}}}")

	log.Tracef("Making request to: %s %s\n", url, string(body))
	request, err = http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", accessToken)

	client := &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 202 {
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Could not create service instance: [%d] %s", response.StatusCode, string(body))
	}

	// Pool job
	job, err = PollJob(cliConnection, response.Header.Get("Location"))
	if err != nil {
		return nil, err
	}

	// Get link to service instance from job
	if link, ok = job.Links["service_instances"]; !ok {
		return nil, fmt.Errorf("Malformed job resource. No 'service_instances' link")
	}

	// Get service instance
	serviceInstance, err = GetServiceInstanceByUrl(cliConnection, *link.Href)

	return &serviceInstance, nil
}
