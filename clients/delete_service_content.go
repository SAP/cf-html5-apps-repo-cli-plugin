package clients

import (
	"cf-html5-apps-repo-cli-plugin/log"
	"errors"
	"net/http"
)

// DeleteServiceContent list HTML5 application files
func DeleteServiceContent(serviceURL string, accessToken string) error {
	var request *http.Request
	var response *http.Response
	var err error
	var html5URL string

	html5URL = serviceURL + "/applications/content"

	log.Tracef("Making request to: %s\n", html5URL)

	client, err := GetDefaultClient()
	if err != nil {
		return err
	}
	if request, err = http.NewRequest("DELETE", html5URL, nil); err != nil {
		return err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	if response, err = client.Do(request); err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}

	return nil
}
