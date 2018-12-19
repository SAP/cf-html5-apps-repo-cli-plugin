package models

// HTML5ListApplicationsResponse response of list applications API
type HTML5ListApplicationsResponse []HTML5App

// HTML5App HTML5 application
type HTML5App struct {
	ApplicationName    string `json:"applicationName,omitempty"`
	ApplicationVersion string `json:"applicationVersion,omitempty"`
	ChangedOn          string `json:"changedOn,omitempty"`
	CreatedOn          string `json:"createdOn,omitempty"`
	IsDefault          bool   `json:"isDefault,omitempty"`
}
