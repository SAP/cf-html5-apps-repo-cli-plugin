package clients

import (
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

// GetToken get token
func GetToken(credentials models.CFCredentials) (string, error) {
	var token string
	var httpClient *http.Client
	var certificate tls.Certificate
	var response *http.Response
	var err error
	var uaaURL string
	var body []byte

	if credentials.UAA.CredentialType == "x509" {
		uaaURL = credentials.UAA.CertURL + "/oauth/token"

		log.Tracef("Making mTLS request to: %s\n", uaaURL)

		certificate, err = tls.X509KeyPair([]byte(credentials.UAA.Certificate), []byte(credentials.UAA.Key))
		if err != nil {
			return "", err
		}

		httpClient, err = GetClientWithCertificates([]tls.Certificate{certificate})
		if err != nil {
			return "", err
		}
		response, err = httpClient.PostForm(uaaURL,
			url.Values{
				"client_id":     {credentials.UAA.ClientID},
				"grant_type":    {"client_credentials"},
				"response_type": {"token"}})
	} else {
		uaaURL = credentials.UAA.URL + "/oauth/token"

		log.Tracef("Making request to: %s\n", uaaURL)

		response, err = http.PostForm(uaaURL,
			url.Values{
				"client_id":     {credentials.UAA.ClientID},
				"client_secret": {credentials.UAA.ClientSecret},
				"grant_type":    {"client_credentials"},
				"response_type": {"token"}})

	}
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
