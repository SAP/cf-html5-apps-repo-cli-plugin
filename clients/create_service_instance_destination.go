package clients

import (
	"bytes"
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"fmt"
	"io"
	"net/http"
)

// CreateServiceInstanceDestination create destination service service instance destination
func CreateServiceInstanceDestination(serviceURL string, accessToken string, destination models.DestinationConfiguration) error {
	var err error
	var request *http.Request
	var response *http.Response
	var destinationsURL string
	var payload []byte
	var body []byte

	log.Tracef("Marshaling destination configuration: %+v\n", destination)
	payload, err = destination.MarshalJSON()
	if err != nil {
		return err
	}
	log.Tracef("Destination configuration JSON: %s\n", payload)

	destinationsURL = serviceURL + "/destination-configuration/v1/instanceDestinations/"
	log.Tracef("Making request to: %s\n", destinationsURL)
	request, err = http.NewRequest("POST", destinationsURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+accessToken)

	client, err := GetDefaultClient()
	if err != nil {
		return err
	}
	response, err = client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err = io.ReadAll(response.Body)
	log.Trace(log.Response{Head: response, Body: body})
	if err != nil {
		return err
	}

	if response.StatusCode > 201 {
		return fmt.Errorf("Could not create destination: [%s] %+v", response.Status, body)
	}

	return nil
}
