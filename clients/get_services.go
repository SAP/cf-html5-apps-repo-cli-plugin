package clients

import (
	"cf-html5-apps-repo-cli-plugin/cache"
	models "cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

// GetServices get Cloud Foundry services
func GetServices(cliConnection plugin.CliConnection) ([]models.CFService, error) {
	var services []models.CFService
	var responseObject models.CFResponse
	var responseStrings []string
	var err error
	var nextURL *string
	var pathStart int
	var pathSlice string

	space, err := cliConnection.GetCurrentSpace()
	if err != nil {
		return nil, err
	}

	if cachedServices, ok := cache.Get("GetServices:" + space.Guid); ok {
		log.Tracef("Returning cached list of services\n")
		services = cachedServices.([]models.CFService)
		return services, nil
	}

	services = make([]models.CFService, 0)
	firstURL := "/v3/service_offerings?space_guids=" + space.Guid
	nextURL = &firstURL

	for nextURL != nil {
		log.Tracef("Making request to: %s\n", *nextURL)
		responseStrings, err = cliConnection.CliCommandWithoutTerminalOutput("curl", *nextURL)
		if err != nil {
			return nil, err
		}

		responseObject = models.CFResponse{}
		body := []byte(strings.Join(responseStrings, ""))
		log.Trace(log.Response{Body: body})
		err = json.Unmarshal(body, &responseObject)
		if err != nil {
			return nil, err
		}

		if len(responseObject.Resources) == 0 {
			log.Tracef("Unexpected response from %q (no resources): %s\n", *nextURL, string(body[:]))
		}

		for _, service := range responseObject.Resources {
			services = append(services, models.CFService{
				Name: service.Name,
				GUID: service.GUID,
			})
		}

		if responseObject.Pagination.Next.Href != nil && *responseObject.Pagination.Next.Href == *nextURL {
			log.Tracef("Unexpected value of the next page URL (equal to previous): %s\n", *nextURL)
			break
		}

		nextURL = responseObject.Pagination.Next.Href
		if nextURL != nil {
			pathStart = strings.Index(*nextURL, "/v3/service_offerings")
			if pathStart > 0 {
				pathSlice = (*nextURL)[pathStart:]
				nextURL = &pathSlice
			}
		}
	}

	log.Tracef("Updating cache with %d service offerings\n", len(services))
	cache.Set("GetServices:"+space.Guid, services)

	return services, nil
}
