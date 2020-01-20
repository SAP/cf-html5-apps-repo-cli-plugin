package models

import "encoding/json"

// DestinationListDestinationsResponse destination service list of destination configurations
type DestinationListDestinationsResponse = []DestinationConfiguration

// DestinationConfiguration destination configuration object
type DestinationConfiguration struct {
	Name                string `json:"Name,omitempty"`
	Description         string `json:"Description,omitempty"`
	Type                string `json:"Type,omitempty"`
	URL                 string `json:"URL,omitempty"`
	Authentication      string `json:"Authentication,omitempty"`
	ProxyType           string `json:"ProxyType,omitempty"`
	TokenServiceURL     string `json:"tokenServiceURL,omitempty"`
	TokenServiceURLType string `json:"tokenServiceURLType,omitempty"`
	ClientID            string `json:"clientId,omitempty"`
	ClientSecret        string `json:"clientSecret,omitempty"`
	Properties          map[string]string
}

// MarshalJSON marshals destination configuration
func (dc *DestinationConfiguration) MarshalJSON() ([]byte, error) {
	jsonMap := make(map[string]string)
	jsonMap["Name"] = dc.Name
	jsonMap["Description"] = dc.Description
	jsonMap["Type"] = dc.Type
	jsonMap["URL"] = dc.URL
	jsonMap["Authentication"] = dc.Authentication
	jsonMap["ProxyType"] = dc.ProxyType
	jsonMap["tokenServiceURL"] = dc.TokenServiceURL
	jsonMap["tokenServiceURLType"] = dc.TokenServiceURLType
	jsonMap["clientId"] = dc.ClientID
	jsonMap["clientSecret"] = dc.ClientSecret
	for key, value := range dc.Properties {
		jsonMap[key] = value
	}
	return json.Marshal(jsonMap)
}

// UnmarshalJSON unmarshals destination configuration
func (dc *DestinationConfiguration) UnmarshalJSON(data []byte) error {
	jsonMap := make(map[string]string)
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return err
	}
	for key, value := range jsonMap {
		switch key {
		case "Name":
			dc.Name = value
		case "Description":
			dc.Description = value
		case "Type":
			dc.Type = value
		case "URL":
			dc.URL = value
		case "Authentication":
			dc.Authentication = value
		case "tokenServiceURL":
			dc.TokenServiceURL = value
		case "TokenServiceURLType":
			dc.TokenServiceURLType = value
		case "clientId":
			dc.ClientID = value
		case "clientSecret":
			dc.ClientSecret = value
		default:
			if dc.Properties == nil {
				dc.Properties = make(map[string]string)
			}
			dc.Properties[key] = value
		}
	}
	return nil
}
