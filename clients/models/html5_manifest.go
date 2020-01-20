package models

// HTML5Manifest HTML5 application manifest file
type HTML5Manifest struct {
	SapApp   HTML5ManifestSapApp   `json:"sap.app,omitempty"`
	SapCloud HTML5ManifestSapCloud `json:"sap.cloud,omitempty"`
}

// HTML5ManifestSapApp HTML5 application manifest file "sap.app" namespace
type HTML5ManifestSapApp struct {
	ID                 string                                `json:"id,omitempty"`
	Type               string                                `json:"type,omitempty"`
	ApplicationVersion HTML5ManifestSapAppApplicationVersion `json:"applicationVersion,omitempty"`
}

// HTML5ManifestSapAppApplicationVersion HTML5 application manifest file "sap.app/applicationVersion" namespace
type HTML5ManifestSapAppApplicationVersion struct {
	Version string `json:"version,omitempty"`
}

// HTML5ManifestSapCloud HTML5 application manifest file "sap.cloud" namespace
type HTML5ManifestSapCloud struct {
	Public  bool   `json:"public,omitempty"`
	Service string `json:"service,omitempty"`
}
