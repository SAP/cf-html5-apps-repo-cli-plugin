package models

// CFErrorResponse Cloud Foundry error response
type CFErrorResponse struct {
	Description *string              `json:"description,omitempty"`
	ErrorCode   *string              `json:"error_code,omitempty"`
	Code        int                  `json:"code,omitempty"`
	HTTP        *CFErrorResponseHTTP `json:"http,omitempty"`
}

// CFErrorResponseHTTP Cloud Foundry error response HTTP data
type CFErrorResponseHTTP struct {
	URI    string `json:"uri,omitempty"`
	Method string `json:"method,omitempty"`
	Status int    `json:"status,omitempty"`
}
