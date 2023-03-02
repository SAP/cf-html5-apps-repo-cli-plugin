package models

// CFErrorResponse Cloud Foundry error response
type CFErrorResponse []CFErrorResponseItem

type CFErrorResponseItem struct {
	Code   int    `json:"code,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
}
