package models

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

// CFCredentials Cloud Foundry response resource entity credentials
type CFCredentials struct {
	Vendor               *string `json:"vendor,omitempty"`
	URI                  *string `json:"uri,omitempty"`
	GrantType            *string `json:"grant_type,omitempty"`
	SapCloudService      *string `json:"sap.cloud.service,omitempty"`
	SapCloudServiceAlias *string `json:"sap.cloud.service.alias,omitempty"`
	UAA                  *CFUAA  `json:"uaa,omitempty"`
}

// CFUAA Cloud Foundry XSUAA credentials
type CFUAA struct {
	ClientID     string `json:"clientid,omitempty"`
	ClientSecret string `json:"clientsecret,omitempty"`
	URL          string `json:"url,omitempty"`
}

// CFLastOperation Cloud Foundry last operation
type CFLastOperation struct {
	Type        string `json:"type,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}
