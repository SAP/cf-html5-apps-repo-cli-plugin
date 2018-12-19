package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

// GetToken get token
func GetToken(credentials models.CFCredentials) (string, error) {
	var token string
	var response *http.Response
	var err error
	var uaaURL string
	var body []byte

	uaaURL = credentials.UAA.URL + "/oauth/token"

	log.Tracef("Making request to: %s\n", uaaURL)

	response, err = http.PostForm(uaaURL,
		url.Values{
			"client_id":     {credentials.UAA.ClientID},
			"client_secret": {credentials.UAA.ClientSecret},
			"grant_type":    {"client_credentials"},
			"response_type": {"token"}})
	if err != nil {
		return "", err
	}

	// Get response body
	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Parse response JSON
	var uaaResponse models.UAAResponse
	err = json.Unmarshal(body, &uaaResponse)
	if err != nil {
		return "", err
	}
	token = uaaResponse.AccessToken

	return token, nil
}
