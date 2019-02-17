package clients

import (
	"bytes"
	"cf-html5-apps-repo-cli-plugin/log"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
)

// UploadAppHost upload ZIP files with HTML5 applications to html5-apps-repo service
func UploadAppHost(serviceURL string, zipFiles []string, accessToken string) error {
	var html5URL string
	var err error

	html5URL = serviceURL + "/applications/content/"

	// Multipart body writer
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add zip files as parts
	for _, zipFile := range zipFiles {
		log.Tracef("Adding '%s' as part to multipart request\n", zipFile)
		// Read application archive
		file, err := os.Open(zipFile)
		if err != nil {
			return err
		}
		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			return err
		}
		file.Close()

		// Write file contents as part
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "apps", fi.Name()))
		h.Set("Content-Type", "application/zip")
		part, err := writer.CreatePart(h)
		if err != nil {
			return err
		}
		part.Write(fileContents)
	}

	// Close request body
	err = writer.Close()
	if err != nil {
		return err
	}

	// Make request
	log.Tracef("Making request to: %s\n", html5URL)
	client := &http.Client{}
	request, err := http.NewRequest("PUT", html5URL, body)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", "Bearer "+accessToken)
	request.Header.Add("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == 201 {
		log.Tracef("Successfully uploaded: %+v\n", zipFiles)
	} else {
		// Get response body
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		bodyString := string(body)
		log.Tracef("Could not upload files: %+v. Response: [%d] %s\n", zipFiles, response.StatusCode, bodyString)
		idx := strings.LastIndex(bodyString, ":")
		// Handle client errors (HTTP 400)
		if response.StatusCode == 400 && idx >= 0 {
			bodyString = bodyString[idx+1:]
			return fmt.Errorf(bodyString)
		}
		// Return error
		return fmt.Errorf("[%d] %s", response.StatusCode, bodyString)
	}

	return nil
}
