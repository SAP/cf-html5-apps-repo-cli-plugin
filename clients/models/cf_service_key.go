package models

// CFServiceKey Cloud Foundry service
type CFServiceKey struct {
	Name        string
	GUID        string
	Credentials CFCredentials
}
