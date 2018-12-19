package models

// HTML5ListApplicationFilesResponse response of list application files API
type HTML5ListApplicationFilesResponse []HTML5ApplicationFile

// HTML5ApplicationFile application file
type HTML5ApplicationFile struct {
	FilePath     string `json:"filePath,omitempty"`
	FileMetadata HTML5ApplicationFileMetadata
}
