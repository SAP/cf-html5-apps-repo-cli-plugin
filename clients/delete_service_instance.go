package clients

import (
	"cf-html5-apps-repo-cli-plugin/log"

	"github.com/cloudfoundry/cli/plugin"
)

// DeleteServiceInstance delete Cloud Foundry service instance
func DeleteServiceInstance(cliConnection plugin.CliConnection, serviceInstanceGUID string) error {
	var err error
	var url string

	url = "/v2/service_instances/" + serviceInstanceGUID + "?recursive=true"
	log.Tracef("Making request to: %s\n", url)
	_, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "DELETE")

	return err
}
