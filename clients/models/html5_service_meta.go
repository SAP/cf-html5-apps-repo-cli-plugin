package models

// HTML5ServiceMeta HTML5 service metadata
type HTML5ServiceMeta struct {
	Status    string `json:"status,omitempty"`
	Size      int    `json:"size,omitempty"`
	SizeLimit int    `json:"sizeLimit,omitempty"`
	ChangedOn string `json:"changedOn,omitempty"`
	Error     error
}
