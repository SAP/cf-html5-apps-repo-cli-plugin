package models

type CFLink struct {
	Href *string `json:"href"`
}

type CFJobMessage struct {
	Code   int64  `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type CFJob struct {
	GUID      string            `json:"guid"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Operation string            `json:"operation"`
	State     string            `json:"state"`
	Errors    []CFJobMessage    `json:"errors"`
	Warnings  []CFJobMessage    `json:"warnings"`
	Links     map[string]CFLink `json:"links"`
}
