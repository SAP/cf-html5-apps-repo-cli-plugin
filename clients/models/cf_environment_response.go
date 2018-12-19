package models

// CFEnvironmentResponse response of 'cf env'
type CFEnvironmentResponse struct {
	SystemEnvJSON      CFSystemEnvJSON      `json:"system_env_json,omitempty"`
	ApplicationEnvJSON CFApplicationEnvJSON `json:"application_env_json,omitempty"`
}

// CFSystemEnvJSON struct
type CFSystemEnvJSON struct {
	VCAPServices map[string][]CFServiceBinding `json:"VCAP_SERVICES,omitempty"`
}

// CFApplicationEnvJSON struct
type CFApplicationEnvJSON struct {
	VCAPApplication map[string]interface{} `json:"VCAP_APPLICATION,omitempty"`
}

// CFServiceBinding struct
type CFServiceBinding struct {
	Name        string               `json:"name,omitempty"`
	Plan        string               `json:"plan,omitempty"`
	Credentials CFBindingCredentials `json:"credentials,omitempty"`
}

// CFBindingCredentials struct
type CFBindingCredentials struct {
	Endpoints       *map[string]string `json:"endpoints,omitempty"`
	HTML5AppsRepo   *HTML5AppsRepo     `json:"html5-apps-repo,omitempty"`
	SAPCloudService *string            `json:"sap.cloud.service,omitempty"`
	UAA             *XSUAA             `json:"uaa,omitempty"`
}

// HTML5AppsRepo struct
type HTML5AppsRepo struct {
	AppHostID string `json:"app_host_id,omitempty"`
}

// XSUAA struct
type XSUAA struct {
	ClientID     string `json:"clientid,omitempty"`
	ClientSecret string `json:"clientsecret,omitempty"`
	URL          string `json:"url,omitempty"`
}
