package models

import "encoding/json"

// CFResponse Cloud Foundry response
type CFResponse struct {
	TotalResults int          `json:"total_results"`
	TotalPages   int          `json:"total_pages"`
	PrevURL      *string      `json:"prev_url,omitempty"`
	NextURL      *string      `json:"next_url,omitempty"`
	Resources    []CFResource `json:"resources"`
}

// CFResource Cloud Foundry response resource
type CFResource struct {
	// metadata
	Metadata *CFResourceMetadata `json:"metadata,omitempty"`

	// entity
	Entity *CFResourceEntity `json:"entity,omitempty"`
}

// CFResourceMetadata Cloud Foundry response resource metadata
type CFResourceMetadata struct {

	// created at
	CreatedAt string `json:"created_at,omitempty"`

	// guid
	GUID string `json:"guid,omitempty"`

	// updated at
	UpdatedAt string `json:"updated_at,omitempty"`

	// url
	URL string `json:"url,omitempty"`
}

// CFResourceEntity Cloud Foundry response resource entity
type CFResourceEntity struct {

	// name
	Name *string `json:"name,omitempty"`

	// label
	Label *string `json:"label,omitempty"`

	// credentials
	Credentials *CFCredentials `json:"credentials,omitempty"`

	// last operation
	LastOperation *CFLastOperation `json:"last_operation,omitempty"`
}

// CFEndpoint business service endpoint
type CFEndpoint struct {
	Timeout string `json:"timeout,omitempty"`
	URL     string `json:"url,omitempty"`
}

// CFCredentials Cloud Foundry response resource entity credentials
type CFCredentials struct {
	Vendor               *string                `json:"vendor,omitempty"`
	URI                  *string                `json:"uri,omitempty"`
	GrantType            *string                `json:"grant_type,omitempty"`
	SapCloudService      *string                `json:"sap.cloud.service,omitempty"`
	SapCloudServiceAlias *string                `json:"sap.cloud.service.alias,omitempty"`
	UAA                  *CFUAA                 `json:"uaa,omitempty"`
	HTML5AppsRepo        *HTML5AppsRepo         `json:"html5-apps-repo,omitempty"`
	Endpoints            *map[string]CFEndpoint `json:"endpoints,omitempty"`
}

// UnmarshalJSON unmarshals Cloud Foundry service credentials
func (credentials *CFCredentials) UnmarshalJSON(data []byte) error {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return err
	}
	for key, value := range jsonMap {
		switch v := value.(type) {
		case string:
			switch key {
			case "vendor":
				credentials.Vendor = &v
			case "uri":
				credentials.URI = &v
			case "grant_type":
				credentials.GrantType = &v
			case "sap.cloud.service":
				credentials.SapCloudService = &v
			case "sap.cloud.service.alias":
				credentials.SapCloudServiceAlias = &v
			case "clientid":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.ClientID = v
			case "clientsecret":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.ClientSecret = v
			case "identityzone":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.IdentityZone = v
			case "url":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.URL = v
			case "xsappname":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.XSAPPNAME = v
			}
		case map[string]interface{}:
			switch key {
			case "uaa":
				credentials.UAA = &CFUAA{}
				for uaaKey, uaaValue := range v {
					switch vv := uaaValue.(type) {
					case string:
						switch uaaKey {
						case "clientid":
							credentials.UAA.ClientID = vv
						case "clientsecret":
							credentials.UAA.ClientSecret = vv
						case "identityzone":
							credentials.UAA.IdentityZone = vv
						case "url":
							credentials.UAA.URL = vv
						case "xsappname":
							credentials.UAA.XSAPPNAME = vv
						}
					}
				}
			case "html5-apps-repo":
				credentials.HTML5AppsRepo = &HTML5AppsRepo{}
				for html5Key, html5Value := range v {
					switch vv := html5Value.(type) {
					case string:
						switch html5Key {
						case "app_host_id":
							credentials.HTML5AppsRepo.AppHostID = vv
						}
					}
				}
			case "endpoints":
				endpoints := make(map[string]CFEndpoint)
				credentials.Endpoints = &endpoints
				for endpointsKey, endpointsValue := range v {
					switch vv := endpointsValue.(type) {
					case string:
						endpoints[endpointsKey] = CFEndpoint{URL: vv, Timeout: ""}
					case map[string]string:
						endpoints[endpointsKey] = CFEndpoint{URL: vv["url"], Timeout: vv["timeout"]}
					}
				}
			}
		}
	}
	return nil
}

// CFUAA Cloud Foundry XSUAA credentials
type CFUAA struct {
	ClientID     string `json:"clientid,omitempty"`
	ClientSecret string `json:"clientsecret,omitempty"`
	IdentityZone string `json:"identityzone,omitempty"`
	URL          string `json:"url,omitempty"`
	XSAPPNAME    string `json:"xsappname,omitempty"`
}

// CFLastOperation Cloud Foundry last operation
type CFLastOperation struct {
	Type        string `json:"type,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}
