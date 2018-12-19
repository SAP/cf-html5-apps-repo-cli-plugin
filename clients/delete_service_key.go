package clients

import (
	"cf-html5-apps-repo-cli-plugin/log"

	"github.com/cloudfoundry/cli/plugin"
)

// DeleteServiceKey delete Cloud Foundry service key
func DeleteServiceKey(cliConnection plugin.CliConnection, serviceKeyGUID string) error {
	var err error
	var url string

	url = "/v2/service_keys/" + serviceKeyGUID
	log.Tracef("Making request to: %s\n", url)
	_, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "DELETE")

	return err
}
