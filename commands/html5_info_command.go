package commands

import (
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"flag"
	"regexp"
	"strings"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
)

// InfoCommand get service instance information
// dependent service keys
type InfoCommand struct {
	HTML5Command
}

// GetPluginCommand returns the plugin command details
func (c *InfoCommand) GetPluginCommand() plugin.Command {
	return plugin.Command{
		Name:     "html5-info",
		HelpText: "Get size limit and status of app-host service instances",
		UsageDetails: plugin.Usage{
			Usage: "cf html5-info [APP_HOST_ID|-n APP_HOST_NAME ...]",
			Options: map[string]string{
				"-name,-n":      "Use app-host service instance with specified name",
				"APP_HOST_ID":   "GUID of html5-apps-repo app-host service instance",
				"APP_HOST_NAME": "Name of html5-apps-repo app-host service instance",
			},
		},
	}
}

// Execute executes plugin command
func (c *InfoCommand) Execute(args []string) ExecutionStatus {
	log.Tracef("Executing command '%s': args: '%v'\n", c.Name, args)

	flagSet := flag.NewFlagSet("html5-info", flag.ContinueOnError)
	var appHostNames stringSlice
	flagSet.Var(&appHostNames, "name", "Name of html5-apps-repo app-host service instance")
	flagSet.Var(&appHostNames, "n", "Name of html5-apps-repo app-host service instance (alias)")
	flagSet.Parse(args)

	appHostGUIDs := flagSet.Args()
	return c.GetServiceInfos(appHostGUIDs, appHostNames)
}

// GetServiceInfos get html5-apps-repo service app-host plan info
func (c *InfoCommand) GetServiceInfos(appHostGUIDs []string, appHostNames []string) ExecutionStatus {
	log.Tracef("Getting information about service instances: %v and %v\n", appHostGUIDs, appHostNames)
	var err error

	// Channel to control number of concurrent connections
	rateLimiter := make(chan int, maxConcurrentConnections)

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	// If no app-host ID passed, get all
	if len(appHostGUIDs) == 0 && len(appHostNames) == 0 {
		ui.Say("Getting information about all app-host service instances in org %s / space %s as %s...",
			terminal.EntityNameColor(context.Org),
			terminal.EntityNameColor(context.Space),
			terminal.EntityNameColor(context.Username))
	} else {

		if len(appHostNames) != 0 {
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
		}

		ui.Say("Getting information about app-host service instances %s in org %s / space %s as %s...",
			terminal.EntityNameColor(strings.Join(appHostGUIDs, ", ")),
			terminal.EntityNameColor(context.Org),
			terminal.EntityNameColor(context.Space),
			terminal.EntityNameColor(context.Username))
	}

	// Get HTML5 context
	html5Context, err := c.GetHTML5Context(context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	// Find app-host service plan
	log.Tracef("Looking for app-host service plan\n")
	var appHostServicePlan *models.CFServicePlan
	for _, plan := range html5Context.HTML5AppsRepoServicePlans {
		if plan.Name == "app-host" {
			appHostServicePlan = &plan
			break
		}
	}
	if appHostServicePlan == nil {
		ui.Failed("Could not find app-host service plan")
		return Failure
	}

	// Get list of service instances of app-host plan
	log.Tracef("Getting service instances of %s service app-host plan (%+v)\n", html5Context.ServiceName, appHostServicePlan)
	var appHostServiceInstances []models.CFServiceInstance
	appHostServiceInstances, err = clients.GetServiceInstances(c.CliConnection, context.SpaceID, []models.CFServicePlan{*appHostServicePlan})
	if err != nil {
		ui.Failed("Could not get service instances for app-host plan: %+v", err)
		return Failure
	}

	// If no app-host ID passed, get all of them.
	// Otherwise find and normalize GUID/Name
	nameMap := make(map[string]string)
	if len(appHostGUIDs) == 0 {
		// Collect app-host IDs
		for _, serviceInstance := range appHostServiceInstances {
			appHostGUIDs = append(appHostGUIDs, serviceInstance.GUID)
			nameMap[serviceInstance.GUID] = serviceInstance.Name
		}
	} else {
		log.Tracef("appHostGUIDs before normalization %+v\n", appHostGUIDs)
		for _, serviceInstance := range appHostServiceInstances {
			log.Tracef("Service id and name '%s'->'%s'\n", serviceInstance.GUID, serviceInstance.Name)
			idx := indexOfString(appHostGUIDs, serviceInstance.Name)
			if idx >= 0 {
				// Name was passed instead of GUID -> change name to GUID
				log.Tracef("Converting service name to id '%s'->'%s'\n", appHostGUIDs[idx], serviceInstance.GUID)
				appHostGUIDs = replaceString(appHostGUIDs, idx, serviceInstance.GUID)
			}
			nameMap[serviceInstance.GUID] = serviceInstance.Name
		}
		// Check that appHostGUIDs are GUIDs
		match := false
		for _, appHostGUID := range appHostGUIDs {
			match, err = regexp.MatchString("^[A-Za-z0-9]{8}-([A-Za-z0-9]{4}-){3}[A-Za-z0-9]{12}$", appHostGUID)
			if err != nil {
				ui.Failed("Regular expression check failed: %+v", err)
				return Failure
			}
			if !match {
				ui.Failed("Argument '%s' is neither existing service instance name, nor app-host-id", appHostGUID)
				return Failure
			}
		}
		log.Tracef("appHostGUIDs after normalization %+v\n", appHostGUIDs)
	}

	sizeMap := make(map[string]int)
	infoMap := make(map[string]models.HTML5ServiceMeta)
	for _, appHostGUID := range appHostGUIDs {
		sizeMap[appHostGUID] = 0

		// Create service key for DT
		log.Tracef("Creating service key for app-host-id '%s'\n", appHostGUID)
		serviceKey, err := clients.CreateServiceKey(c.CliConnection, appHostGUID, nil)
		if err != nil {
			ui.Failed("Could not create service key for service instance with id '%s' : %+v", appHostGUID, err)
			return Failure
		}

		// Obtain access token
		log.Tracef("Obtaining access token for service key '%s'\n", serviceKey.Name)
		token, err := clients.GetToken(serviceKey.Credentials)
		if err != nil {
			ui.Failed("Could not obtain access token for service key '%s': %+v", serviceKey.Name, err)
			return Failure
		}
		log.Tracef("Access token for service key '%s': %s\n",
			serviceKey.Name,
			log.Sensitive{Data: token})

		// Get app-host service info
		log.Tracef("Getting information about service with app-host-id '%s'\n", appHostGUID)
		infoChan := make(chan models.HTML5ServiceMeta)
		go clients.GetServiceMeta(*serviceKey.Credentials.URI, token, infoChan)
		info := <-infoChan
		infoMap[appHostGUID] = info

		// Check if app-host has size metadata
		if info.Size > 0 {
			log.Tracef("Service instance '%s' contains size metadata: %d\n", appHostGUID, info.Size)
			sizeMap[appHostGUID] = info.Size
		} else {
			// Fallback to sum of HEAD sizes of all files
			log.Tracef("Service instance '%s' does no contains size metadata\n", appHostGUID)

			// Get list of app-host applications
			apps, err := clients.ListApplicationsForAppHost(
				*html5Context.HTML5AppRuntimeServiceInstanceKeys[len(html5Context.HTML5AppRuntimeServiceInstanceKeys)-1].Credentials.URI,
				html5Context.HTML5AppRuntimeServiceInstanceKeyToken, appHostGUID)
			if err != nil {
				ui.Failed("Could not get list of applications for app-host-id '%s': %+v", appHostGUID, err)
				return Failure
			}

			for _, app := range apps {
				// Get list of application files
				files, err := clients.ListFilesOfApp(*html5Context.HTML5AppRuntimeServiceInstanceKeys[len(html5Context.HTML5AppRuntimeServiceInstanceKeys)-1].Credentials.URI,
					app.ApplicationName+"-"+app.ApplicationVersion,
					html5Context.HTML5AppRuntimeServiceInstanceKeyToken, appHostGUID)
				if err != nil {
					ui.Failed("Could not get list of application files for app-host-id '%s' and application '%s': %+v", appHostGUID,
						app.ApplicationName+"-"+app.ApplicationVersion, err)
					return Failure
				}
				metaChannel := make(chan models.HTML5ApplicationFileMetadata, len(files))
				log.Tracef("Number of files in the app '%s' is '%d'\n", app.ApplicationName, len(files))
				for _, file := range files {
					rateLimiter <- 1
					// Get file size
					go func(file models.HTML5ApplicationFile) {
						clients.GetFileMeta(*html5Context.HTML5AppRuntimeServiceInstanceKeys[len(html5Context.HTML5AppRuntimeServiceInstanceKeys)-1].Credentials.URI, file.FilePath,
							html5Context.HTML5AppRuntimeServiceInstanceKeyToken, appHostGUID, metaChannel)
						<-rateLimiter
					}(file)
				}
				for range files {
					meta := <-metaChannel
					if meta.Error != nil {
						ui.Failed("Could not get file metadata: %+v", err)
						return Failure
					}
					sizeMap[appHostGUID] += meta.FileSize
				}
			}

		}

		// Delete temporarry service keys
		log.Tracef("Deleting temporarry service key: '%s'\n", serviceKey.Name)
		err = clients.DeleteServiceKey(c.CliConnection, serviceKey.GUID, maxRetryCount)
		if err != nil {
			ui.Failed("Could not delete service key '%s' : %+v", serviceKey.Name, err)
			return Failure
		}
	}

	// Extract app-host infos
	infoRecords := make([]InfoRecord, 0)
	for appHostGUID, meta := range infoMap {
		if meta.Error != nil {
			ui.Failed("Could not read information about service with app-host-id '%s' : %+v", appHostGUID, err)
			return Failure
		}
		infoRecords = append(infoRecords, InfoRecord{
			AppHostName: nameMap[appHostGUID],
			AppHostGUID: appHostGUID,
			SizeLimit:   meta.SizeLimit,
			Used:        sizeMap[appHostGUID],
			Status:      meta.Status,
			ChangedOn:   meta.ChangedOn})
	}

	// Clean-up HTML5 context
	err = c.CleanHTML5Context(html5Context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	ui.Ok()
	ui.Say("")

	// Display information about HTML5 applications
	table := ui.Table([]string{"name", "app-host-id", "used", "size limit", "status", "last changed"})
	for _, infoRecord := range infoRecords {
		table.Add(infoRecord.AppHostName,
			infoRecord.AppHostGUID,
			getReadableSize(infoRecord.Used),
			getReadableSize(infoRecord.SizeLimit),
			infoRecord.Status,
			infoRecord.ChangedOn)
	}
	table.Print()

	return Success
}

func replaceString(collection []string, idx int, element string) []string {
	newCollection := collection[0:idx]
	newCollection = append(newCollection, element)
	newCollection = append(newCollection, collection[idx+1:]...)
	return newCollection
}

// InfoRecord service information record
type InfoRecord struct {
	AppHostName string
	AppHostGUID string
	SizeLimit   int
	Used        int
	Status      string
	ChangedOn   string
}
