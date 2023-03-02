package models

import (
	"encoding/json"
	"fmt"
)

// CFResponse Cloud Foundry response
type CFResponse struct {
	Pagination CFPagination `json:"pagination"`
	Resources  []CFResource `json:"resources"`
}

// CFPagination Cloud Foundry resource pagination
type CFPagination struct {
	TotalResults int    `json:"total_results"`
	TotalPages   int    `json:"total_pages"`
	First        CFLink `json:"first"`
	Last         CFLink `json:"last"`
	Next         CFLink `json:"next"`
	Previous     CFLink `json:"previous"`
}

// CFResource Cloud Foundry response resource
type CFResource struct {
	GUID             string                    `json:"guid"`
	CreatedAt        string                    `json:"created_at"`
	UpdatedAt        string                    `json:"updated_at"`
	Name             string                    `json:"name"`
	Tags             []string                  `json:"tags"`
	Type             string                    `json:"type"`
	MaintenanceInfo  CFMaintenanceInfo         `json:"maintenance_info"`
	IpgradeAvailable bool                      `json:"upgrade_available"`
	DashboardUrl     string                    `json:"dashboard_url"`
	LastOperation    CFLastOperation           `json:"last_operation"`
	Relationships    map[string]CFRelationship `json:"relationships"`
	Metadata         CFMetadata                `json:"metadata"`
	Links            map[string]CFLink         `json:"links"`
}

// CFMaintenanceInfo Cloud Foudry response maintenance info
type CFMaintenanceInfo struct {
	Version string `json:"version"`
}

// CFRelationship Cloud Foudry response resource relationship
type CFRelationship struct {
	Data CFRelationshipData `json:"data"`
}

// CFRelationshipData Cloud Foudry response resource relationship data
type CFRelationshipData struct {
	GUID string `json:"guid"`
}

// CFMetadata Cloud Foundry resource metadata
type CFMetadata struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
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
			case "certificate":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.Certificate = v
			case "certurl":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.CertURL = v
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
			case "credential-type":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.CredentialType = v
			case "identityzone":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.IdentityZone = v
			case "key":
				if credentials.UAA == nil {
					credentials.UAA = &CFUAA{}
				}
				credentials.UAA.Key = v
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
						case "certificate":
							credentials.UAA.Certificate = vv
						case "certurl":
							credentials.UAA.CertURL = vv
						case "clientid":
							credentials.UAA.ClientID = vv
						case "clientsecret":
							credentials.UAA.ClientSecret = vv
						case "credential-type":
							credentials.UAA.CredentialType = vv
						case "identityzone":
							credentials.UAA.IdentityZone = vv
						case "key":
							credentials.UAA.Key = vv
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
					case map[string]interface{}:
						var url string
						var timeout string
						for endpointKey, endpointValue := range vv {
							switch vvv := endpointValue.(type) {
							case string:
								if endpointKey == "url" {
									url = vvv
								}
							case float64:
								if endpointKey == "timeout" {
									timeout = fmt.Sprintf("%g", vvv)
								}
							}
						}
						endpoints[endpointsKey] = CFEndpoint{URL: url, Timeout: timeout}
					}
				}
			}
		}
	}
	return nil
}

// CFUAA Cloud Foundry XSUAA credentials
type CFUAA struct {
	Certificate    string `json:"certificate,omitempty"`
	CertURL        string `json:"certurl,omitempty"`
	ClientID       string `json:"clientid,omitempty"`
	ClientSecret   string `json:"clientsecret,omitempty"`
	CredentialType string `json:"credential-type,omitempty"`
	IdentityZone   string `json:"identityzone,omitempty"`
	Key            string `json:"key,omitempty"`
	URL            string `json:"url,omitempty"`
	XSAPPNAME      string `json:"xsappname,omitempty"`
}

// CFLastOperation Cloud Foundry last operation
type CFLastOperation struct {
	Type        string `json:"type,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}
