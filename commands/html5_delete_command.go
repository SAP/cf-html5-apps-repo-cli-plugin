package commands

import (
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"flag"
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
		HelpText: "Delete one or multiple app-host service instances or content uploaded with these instances",
		UsageDetails: plugin.Usage{
			Usage: "cf html5-delete [--content|--destination] APP_HOST_ID|-n APP_HOST_NAME [...]",
			Options: map[string]string{
				"-content":        "delete content only",
				"-destination,-d": "delete destinations that point to service instances to be deleted",
				"-name,-n":        "Use app-host service instance with specified name",
				"APP_HOST_ID":     "GUID of html5-apps-repo app-host service instance",
				"APP_HOST_NAME":   "Name of html5-apps-repo app-host service instance",
			},
		},
	}
}

// Execute executes plugin command
func (c *DeleteCommand) Execute(args []string) ExecutionStatus {
	log.Tracef("Executing command '%s': args: '%v'\n", c.Name, args)

	flagSet := flag.NewFlagSet("html5-delete", flag.ContinueOnError)
	contentFlag := flagSet.Bool("content", false, "delete content only")
	destinationFlag := flagSet.Bool("destination", false, "delete destinations that point to service instances to be deleted")
	destinationFlagAlias := flagSet.Bool("d", false, "delete destinations that point to service instances to be deleted")

	var appHostNames stringSlice
	flagSet.Var(&appHostNames, "name", "Name of html5-apps-repo app-host service instance")
	flagSet.Var(&appHostNames, "n", "Name of html5-apps-repo app-host service instance (alias)")
	flagSet.Parse(args)

	// Normalize aliases
	if *destinationFlagAlias && !*destinationFlag {
		destinationFlag = destinationFlagAlias
	}

	if flagSet.NArg() > 0 || len(appHostNames) > 0 {
		appHostGUIDs := flagSet.Args()
		if *contentFlag {
			return c.DeleteServiceInstancesContent(appHostGUIDs, appHostNames)
		}
		return c.DeleteServiceInstances(appHostGUIDs, appHostNames, *destinationFlag)
	}

	ui.Failed("Incorrect number of arguments passed. See [cf html5-delete --help] for more detals")
	return Failure
}

// DeleteServiceInstancesContent delete service instances content by app-host-ids
func (c *DeleteCommand) DeleteServiceInstancesContent(appHostGUIDs []string, appHostNames []string) ExecutionStatus {
	log.Tracef("Deleting content of service instances by app-host-ids: %v\n", appHostGUIDs)
	var err error

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	for _, appHostName := range appHostNames {
		// Resolve app-host-id
		log.Tracef("Resolving app-host-id by service instance name '%s'\n", appHostName)
		serviceInstance, err := clients.GetServiceInstanceByName(c.CliConnection, context.SpaceID, appHostName)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		log.Tracef("Resolved app-host-id is '%s'\n", serviceInstance.GUID)
		appHostGUIDs = append(appHostGUIDs, serviceInstance.GUID)
	}

	ui.Say("Deleting content of service instances with app-host-id %s in org %s / space %s as %s...",
		terminal.EntityNameColor(strings.Join(appHostGUIDs, ", ")),
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

	for _, appHostGUID := range appHostGUIDs {
		// Create service key for DT
		log.Tracef("Creating service key for app-host-id '%s'\n", appHostGUID)
		serviceKey, err := clients.CreateServiceKey(c.CliConnection, appHostGUID)
		if err != nil {
			ui.Failed("Could not create service key for service instance with id '%s' : %+v\n", appHostGUID, err)
			return Failure
		}

		// Obtain access token
		log.Tracef("Obtaining access token for service key '%s'\n", serviceKey.Name)
		token, err := clients.GetToken(serviceKey.Credentials)
		if err != nil {
			ui.Failed("Could not obtain access token for service key '': %+v\n", serviceKey.Name, err)
			return Failure
		}

		// Delete app-host service content
		log.Tracef("Deleting content of service with app-host-id '%s'\n", appHostGUID)
		if clients.DeleteServiceContent(*serviceKey.Credentials.URI, token) != nil {
			ui.Failed("Could not delete content of service with app-host-id '%s' : %+v\n", appHostGUID, err)
			return Failure
		}

		// Delete temporarry service keys
		log.Tracef("Deleting temporarry service key: '%s'\n", serviceKey.Name)
		err = clients.DeleteServiceKey(c.CliConnection, serviceKey.GUID, maxRetryCount)
		if err != nil {
			ui.Failed("Could not delete service key '%s' : %+v\n", serviceKey.Name, err)
			return Failure
		}
	}

	ui.Ok()
	ui.Say("")

	return Success
}

// DeleteServiceInstances delete service instances by app-host-ids,
// including all dependent service keys
func (c *DeleteCommand) DeleteServiceInstances(appHostGUIDs []string, appHostNames []string, deleteDestinations bool) ExecutionStatus {
	log.Tracef("Deleting service instances by IDs: %v\n", appHostGUIDs)
	var err error

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s\n", err.Error())
		return Failure
	}

	for _, appHostName := range appHostNames {
		if appHostName[len(appHostName)-1:] == "*" {
			// Resolve service names by prefix
			log.Tracef("Resolving service instances by name prefix '%s'\n", appHostName)
			serviceInstances, err := clients.GetServiceInstancesByNamePrefix(c.CliConnection, context.SpaceID, appHostName)
			if err != nil {
				ui.Failed("Could not get service instances by name prefix '%s': %+v\n", appHostName, err)
				return Failure
			}
			for _, serviceInstance := range serviceInstances {
				log.Tracef("Resolved service instance for name prefix '%s': %+v\n", appHostName, serviceInstance)
				appHostGUIDs = append(appHostGUIDs, serviceInstance.GUID)
			}
		} else {
			// Resolve app-host-id
			log.Tracef("Resolving app-host-id by service instance name '%s'\n", appHostName)
			serviceInstance, err := clients.GetServiceInstanceByName(c.CliConnection, context.SpaceID, appHostName)
			if err != nil {
				ui.Failed("Could not get service instance by name: %+v\n", err)
				return Failure
			}
			log.Tracef("Resolved app-host-id is '%s'\n", serviceInstance.GUID)
			appHostGUIDs = append(appHostGUIDs, serviceInstance.GUID)
		}
	}

	msg := ""
	if deleteDestinations {
		msg = "and associated destinations "
	}

	ui.Say("Deleting service instances with IDs %s %sin org %s / space %s as %s...",
		terminal.EntityNameColor(strings.Join(appHostGUIDs, ", ")),
		msg,
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

	// Delete destinatons if needed
	if deleteDestinations {
		// Create destination context
		log.Tracef("Getting destination service context\n")
		destinationContext, err := c.GetDestinationContext(context)
		if err != nil {
			ui.Failed("Could not create destination context: %+v\n", err)
			return Failure
		}

		// List destinations
		destinations, err := clients.ListSubaccountDestinations(
			*destinationContext.DestinationServiceInstanceKey.Credentials.URI,
			destinationContext.DestinationServiceInstanceKeyToken)

		// Find relevant destinations and delete them
		sapCloudServices := make([]string, 0)
		for _, destination := range destinations {
			val, ok := destination.Properties["html5-apps-repo.app_host_id"]
			if !ok {
				val, ok = destination.Properties["app_host_id"]
			}
			if ok {
				val = strings.Trim(val, " ")
				for _, appHostGUID := range appHostGUIDs {
					if val == appHostGUID {
						log.Tracef("Deleting destination '%s'\n", destination.Name)
						err = clients.DeleteSubaccountDestination(
							*destinationContext.DestinationServiceInstanceKey.Credentials.URI,
							destinationContext.DestinationServiceInstanceKeyToken,
							destination.Name)
						if err != nil {
							ui.Failed("Could not delete destination '%s': %+v\n", destination.Name, err)
							return Failure
						}
						if sapCloudService, ok := destination.Properties["sap.cloud.service"]; ok {
							log.Tracef("Adding sap.cloud.service '%s' to the deletion list\n", sapCloudService)
							sapCloudServices = append(sapCloudServices, sapCloudService)
						}
					}
				}
			}
		}

		// Find destinations with same sap.cloud.service value as deleted destinations and delete them
		if len(sapCloudServices) > 0 {
			for _, destination := range destinations {
				if val, ok := destination.Properties["sap.cloud.service"]; ok {
					for _, sapCloudService := range sapCloudServices {
						if val == sapCloudService {
							log.Tracef("Deleting destination '%s' with sap.cloud.service '%s'\n", destination.Name, val)
							err = clients.DeleteSubaccountDestination(
								*destinationContext.DestinationServiceInstanceKey.Credentials.URI,
								destinationContext.DestinationServiceInstanceKeyToken,
								destination.Name)
							if err != nil {
								ui.Failed("Could not delete destination '%s' with sap.cloud.service '%s': %+v\n",
									destination.Name, val, err)
								return Failure
							}
							break
						}
					}
				}
			}
		}

		// Clen-up destination context
		err = c.CleanDestinationContext(destinationContext)
		if err != nil {
			ui.Failed("Could not clean destination context: %+v\n", err)
			return Failure
		}
	}

	// Delete service instances
	for _, appHostGUID := range appHostGUIDs {
		log.Tracef("Getting list of service keys for app-host-id %s\n", appHostGUID)
		// Get service keys
		serviceKeys, err := clients.GetServiceKeys(c.CliConnection, appHostGUID)
		if err != nil {
			ui.Failed("Could not get list of service keys for app-host-id %s: %+v\n", appHostGUID, err)
			return Failure
		}
		// Delete dependent service keys
		for _, serviceKey := range serviceKeys {
			log.Tracef("Deleting service key %s (%s)\n", serviceKey.GUID, serviceKey.Name)
			err = clients.DeleteServiceKey(c.CliConnection, serviceKey.GUID, maxRetryCount)
			if err != nil {
				ui.Failed("Could not delete service key %s: %+v\n", serviceKey.GUID, err)
				return Failure
			}
		}
		log.Tracef("Deleting service instance %s\n", appHostGUID)
		// Delete service instance
		err = clients.DeleteServiceInstance(c.CliConnection, appHostGUID, maxRetryCount)
		if err != nil {
			if deleteDestinations {
				log.Tracef("Service instance %s was not deleted (probably not found)\n", appHostGUID)
				ui.Warn("[WARNING] Service instance with ID = %s was not deleted (probably not found)\n", appHostGUID)
			} else {
				ui.Failed("Could not delete service instance %s: %+v\n", appHostGUID, err)
				return Failure
			}
		} else {

		}
	}

	ui.Ok()
	ui.Say("")

	return Success
}
