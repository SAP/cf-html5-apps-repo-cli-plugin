package commands

import (
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"strings"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
)

// DeleteCommand delete service instances and all
// dependent service keys
type DeleteCommand struct {
	HTML5Command
}

// GetPluginCommand returns the plugin command details
func (c *DeleteCommand) GetPluginCommand() plugin.Command {
	return plugin.Command{
		Name:     "html5-delete",
		HelpText: "Delete one or multiple app-host service instances and all dependent service keys",
		UsageDetails: plugin.Usage{
			Usage: "cf html5-delete APP_HOST_ID [...]",
			Options: map[string]string{
				"APP_HOST_ID": "GUID of html5-apps-repo app-host service instance",
			},
		},
	}
}

// Execute executes plugin command
func (c *DeleteCommand) Execute(args []string) ExecutionStatus {
	log.Tracef("Executing command '%s': args: '%v'\n", c.Name, args)

	// Parse arguments
	var key = "_"
	var argsMap = make(map[string][]string)
	for _, arg := range args {
		if string(arg[0]) == "-" {
			key = arg
			continue
		}
		if argsMap[key] == nil {
			argsMap[key] = make([]string, 0)
		}
		argsMap[key] = append(argsMap[key], arg)
		key = "_"
	}

	// Define app-host-id

	if argsMap["_"] != nil {
		// Delete service instances
		return c.DeleteServiceInstances(argsMap["_"])
	}

	ui.Failed("Incorrect number of arguments passed. See [cf html5-delete --help] for more detals")
	return Failure
}

// DeleteServiceInstances delete service instances by app-host-ids,
// including all dependent service keys
func (c *DeleteCommand) DeleteServiceInstances(appHostGUIDs []string) ExecutionStatus {
	log.Tracef("Deleting service instances by app-host-ids: %v\n", appHostGUIDs)
	var err error

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	ui.Say("Deleting service instances with app-host-id %s in org %s / space %s as %s...",
		terminal.EntityNameColor(strings.Join(appHostGUIDs, ", ")),
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

	// Delete service instances
	for _, appHostGUID := range appHostGUIDs {
		log.Tracef("Getting list of service keys for app-host-id %s\n", appHostGUID)
		// Get service keys
		serviceKeys, err := clients.GetServiceKeys(c.CliConnection, appHostGUID)
		if err != nil {
			ui.Failed("Could not get list of service keys for app-host-id %s: %+v", appHostGUID, err)
			return Failure
		}
		// Delete dependent service keys
		for _, serviceKey := range serviceKeys {
			log.Tracef("Deleting service key %s (%s)\n", serviceKey.GUID, serviceKey.Name)
			err = clients.DeleteServiceKey(c.CliConnection, serviceKey.GUID)
			if err != nil {
				ui.Failed("Could not delete service key %s: %+v", serviceKey.GUID, err)
				return Failure
			}
		}
		log.Tracef("Deleting service instance %s\n", appHostGUID)
		// Delete service instance
		err = clients.DeleteServiceInstance(c.CliConnection, appHostGUID)
		if err != nil {
			ui.Failed("Could not delete service instance %s: %+v", appHostGUID, err)
			return Failure
		}
	}

	ui.Ok()
	ui.Say("")

	return Success
}
