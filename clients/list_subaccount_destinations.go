package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ListSubaccountDestinations list destination service subaccount destinations
func ListSubaccountDestinations(serviceURL string, accessToken string) (models.DestinationListDestinationsResponse, error) {
	var destinations models.DestinationListDestinationsResponse
	var request *http.Request
	var response *http.Response
	var err error
	var destinationsURL string
	var body []byte

	destinationsURL = serviceURL + "/destination-configuration/v1/subaccountDestinations/"

	log.Tracef("Making request to: %s\n", destinationsURL)

	client := &http.Client{}
	request, err = http.NewRequest("GET", destinationsURL, nil)
	if err != nil {
		return destinations, err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	response, err = client.Do(request)
	if err != nil {
		return destinations, err
	}

	// Get response body
	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return destinations, err
	}

	// Parse response JSON
	err = json.Unmarshal(body, &destinations)
	if err != nil {
		return destinations, err
	}

	return destinations, nil
}
