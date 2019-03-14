package commands

import (
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"fmt"
	"strconv"

	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
)

// ListCommand prints the list of HTML5 applications
// deployed using multiple instances of html5-apps-repo
// service app-host plan
type ListCommand struct {
	HTML5Command
}

// GetPluginCommand returns the plugin command details
func (c *ListCommand) GetPluginCommand() plugin.Command {
	return plugin.Command{
		Name:     "html5-list",
		HelpText: "Display list of HTML5 applications or file paths of specified application",
		UsageDetails: plugin.Usage{
			Usage: "cf html5-list [APP_NAME] [APP_VERSION] [APP_HOST_ID|-n APP_HOST_NAME] [-a CF_APP_NAME [-u]]",
			Options: map[string]string{
				"APP_NAME":      "Application name, which file paths should be listed. If not provided, list of applications will be printed",
				"APP_VERSION":   "Application version, which file paths should be listed. If not provided, current active version will be used",
				"APP_HOST_ID":   "GUID of html5-apps-repo app-host service instance that contains application with specified name and version",
				"APP_HOST_NAME": "Name of html5-apps-repo app-host service instance that contains application with specified name and version",
				"-name, -n":     "Use html5-apps-repo app-host service instance name instead of APP_HOST_ID",
				"-app, -a":      "Cloud Foundry application name, which is bound to services that expose UI via html5-apps-repo",
				"-url, -u":      "Show conventional URLs of applications, when accessed via Cloud Foundry application specified with --app flag",
			},
		},
	}
}

// Execute executes plugin command
func (c *ListCommand) Execute(args []string) ExecutionStatus {
	log.Tracef("Executing command '%s': args: '%v'\n", c.Name, args)

	// List apps in the space
	if len(args) == 0 {
		return c.ListSpaceApps()
	}

	// Parse arguments
	var key = "_"
	var argsMap = make(map[string][]string)
	for _, arg := range args {
		if string(arg[0]) == "-" {
			key = arg
			if argsMap[key] == nil {
				argsMap[key] = make([]string, 0)
			}
			continue
		}
		argsMap[key] = append(argsMap[key], arg)
		key = "_"
	}

	// Service Name
	var name = ""
	if argsMap["-n"] != nil && argsMap["--name"] != nil {
		ui.Failed("Can't use both '--name' and '-n' at the same time")
		return Failure
	}
	if argsMap["-n"] != nil {
		argsMap["--name"] = argsMap["-n"]
	}
	if argsMap["--name"] != nil {
		if len(argsMap["--name"]) != 1 {
			ui.Failed("Incorrect number of arguments for APP_HOST_NAME option (expected: 1, actual: %d). For help see [cf html5-list --help]", len(argsMap["--name"]))
			return Failure
		}
		if len(argsMap["_"]) != 2 {
			ui.Failed("HTML5 application name and version are required, when using '--name' option")
			return Failure
		}
		name = argsMap["--name"][0]
	}

	// App
	var app = ""
	if argsMap["-a"] != nil && argsMap["--app"] != nil {
		ui.Failed("Can't use both '--app' and '-a' at the same time")
		return Failure
	}
	if argsMap["-a"] != nil {
		argsMap["--app"] = argsMap["-a"]
	}
	if argsMap["--app"] != nil {
		if len(argsMap["--app"]) != 1 {
			ui.Failed("Incorrect number of arguments for CF_APP_NAME option (expected: 1, actual: %d). For help see [cf html5-list --help]", len(argsMap["--app"]))
			return Failure
		}
		app = argsMap["--app"][0]
	}

	// Show URLs
	var showUrls = false
	if argsMap["-u"] != nil || argsMap["--url"] != nil {
		showUrls = true
	}

	if app != "" {
		// List HTML5 applications available in CF application context
		return c.ListAppApps(app, showUrls)
	} else if len(argsMap["_"]) == 3 {
		// List files paths of application with version from app-host-id
		return c.ListAppFiles(argsMap["_"][0], argsMap["_"][1], argsMap["_"][2], false)
	} else if len(argsMap["_"]) == 2 {
		// List files paths of application with version
		return c.ListAppFiles(argsMap["_"][0], argsMap["_"][1], name, name != "")
	} else if len(argsMap["_"]) == 1 {
		// List files paths of application default version
		return c.ListAppFiles(argsMap["_"][0], "", "", false)
	}

	ui.Failed("Too much arguments. See [cf html5-list --help] for more detals")
	return Failure
}

// ListAppApps get list of HTML5 applications available in CF application context
func (c *ListCommand) ListAppApps(appName string, showUrls bool) ExecutionStatus {
	log.Tracef("Listing HTML5 applications available for CF application '%s'\n", appName)

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	ui.Say("Getting list of HTML5 application available in scope of application %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appName),
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

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

	// Get list of applications for each app-host service instance
	var data Model
	data.Services = make([]Service, 0)
	for _, serviceInstance := range appHostServiceInstances {
		log.Tracef("Getting list of applications for app-host plan (%+v)\n", serviceInstance)
		applications, err := clients.ListApplicationsForAppHost(*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
			html5Context.HTML5AppRuntimeServiceInstanceKeyToken, serviceInstance.GUID)
		if err != nil {
			ui.Failed("Could not get list of applications for app-host instance %s: %+v", serviceInstance.Name, err)
			return Failure
		}
		apps := make([]App, 0)
		for _, app := range applications {
			apps = append(apps, App{Name: app.ApplicationName, Version: app.ApplicationVersion, Changed: app.ChangedOn, Public: app.IsPublic})
		}
		data.Services = append(data.Services, Service{Name: serviceInstance.Name, GUID: serviceInstance.GUID, Apps: apps})
	}

	// Get Cloud Foundry application details
	app, err := clients.GetApplication(c.CliConnection, context.SpaceID, appName)
	if err != nil {
		ui.Failed("Could not get application metadata: %s", err.Error())
		return Failure
	}

	// Get Cloud Foundry application environment
	env, err := clients.GetEnvironment(c.CliConnection, app.GUID)
	if err != nil {
		ui.Failed("Could not get application environment: %s", err.Error())
		return Failure
	}

	// Find services with app-host-id
	var servicesData = Model{}
	servicesData.Services = make([]Service, 0)
	for serviceName, serviceBindings := range env.SystemEnvJSON.VCAPServices {
		for _, serviceBinding := range serviceBindings {
			if serviceBinding.Credentials.HTML5AppsRepo != nil {
				AppHostIDs := strings.Split(serviceBinding.Credentials.HTML5AppsRepo.AppHostID, ",")
				prefix := ""
				if serviceBinding.Credentials.SAPCloudServiceAlias != nil {
					prefix = *serviceBinding.Credentials.SAPCloudServiceAlias + "."
				} else if serviceBinding.Credentials.SAPCloudService != nil {
					prefix = strings.Replace(strings.Replace(*serviceBinding.Credentials.SAPCloudService, ".", "", -1), "-", "", -1) + "."
				}
				for _, appHostID := range AppHostIDs {
					// Get list of applications for app-host-id
					log.Tracef("Getting list of applications for service '%s' and app-host-id '%s'\n", serviceName, appHostID)
					applications, err := clients.ListApplicationsForAppHost(*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
						html5Context.HTML5AppRuntimeServiceInstanceKeyToken, appHostID)
					if err != nil {
						ui.Failed("Could not get list of applications for app-host-id '%s': %+v", appHostID, err)
						return Failure
					}
					apps := make([]App, 0)
					for _, app := range applications {
						apps = append(apps, App{Name: app.ApplicationName, Version: app.ApplicationVersion, Changed: app.ChangedOn, Public: app.IsPublic})
					}

					servicesData.Services = append(servicesData.Services, Service{GUID: appHostID, Name: serviceName, Apps: apps, Prefix: prefix})
				}
			}
		}
	}

	// Clean-up HTML5 context
	err = c.CleanHTML5Context(html5Context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	ui.Ok()
	ui.Say("")

	columns := make([]string, 0)
	columns = append(columns, "name", "version", "app-host-id", "service instance", "visibility", "last changed")
	if showUrls {
		columns = append(columns, "url")
	}

	// Display information about HTML5 applications
	table := ui.Table(columns)
	type ColorFunction = func(message string) string
	addRow := func(service Service, app App, fn ColorFunction) {
		row := make([]string, 0)
		row = append(row,
			app.Name,
			fn(app.Version),
			fn(service.GUID),
			fn(service.Name),
			fn((map[bool]string{true: "public", false: "private"})[app.Public]),
			fn(app.Changed))
		if showUrls {
			row = append(row, fn("https://"+env.ApplicationEnvJSON.VCAPApplication.Uris[0]+"/"+service.Prefix+app.Name+"-"+app.Version+"/"))
		}
		table.Add(row...)
	}
	for _, service := range data.Services {
		for _, app := range service.Apps {
			addRow(service, app, terminal.LogStdoutColor)
		}
	}
	for _, service := range servicesData.Services {
		for _, app := range service.Apps {
			addRow(service, app, terminal.AdvisoryColor)
		}
	}
	table.Print()

	return Success
}

// ListAppFiles get list of application files
func (c *ListCommand) ListAppFiles(appName string, appVersion string, appHostNameOrID string, isName bool) ExecutionStatus {
	log.Tracef("Listing application file paths for name '%s': version: '%s'\n", appName, appVersion)

	// Calculate application key
	var appKey = appName
	if appVersion != "" {
		appKey = appName + "-" + appVersion
	}

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	ui.Say("Getting list of files for HTML5 application %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appKey),
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

	// Get HTML5 context
	html5Context, err := c.GetHTML5Context(context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	appHostID := appHostNameOrID
	if isName {
		// Resolve app-host-id
		log.Tracef("Resolving app-host-id by service instance name '%s'\n", appHostNameOrID)
		serviceInstance, err := clients.GetServiceInstanceByName(c.CliConnection, context.SpaceID, appHostNameOrID)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		log.Tracef("Resolved app-host-id is '%s'\n", serviceInstance.GUID)
		appHostID = serviceInstance.GUID
	}

	// Find active version
	if appVersion == "" {
		log.Tracef("Getting list of applications for app-runtime plan (%+v)\n", html5Context.HTML5AppRuntimeServiceInstances[len(html5Context.HTML5AppRuntimeServiceInstances)-1].Name)
		applications, err := clients.ListApplicationsForAppRuntime(*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI, html5Context.HTML5AppRuntimeServiceInstanceKeyToken)
		if err != nil {
			ui.Failed("Could not get list of applications for app-runtime instance %s: %+v", html5Context.HTML5AppRuntimeServiceInstances[len(html5Context.HTML5AppRuntimeServiceInstances)-1].Name, err)
			return Failure
		}
		for _, application := range applications {
			if application.ApplicationName == appName && application.IsDefault {
				appVersion = application.ApplicationVersion
				appKey = appName + "-" + appVersion
				log.Tracef("Default version for application %s is %s\n", appName, appVersion)
				break
			}
		}
	}

	// Get list of files
	files, err := clients.ListFilesOfApp(
		*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
		appKey,
		html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
		appHostID)
	if err != nil {
		ui.Failed("Could not list of files for app %s: %+v", appKey, err)
		return Failure
	}

	// Get files size and etag
	start := time.Now()
	metas := make([]chan models.HTML5ApplicationFileMetadata, len(files))
	for idx := range files {
		metas[idx] = make(chan models.HTML5ApplicationFileMetadata)
		go clients.GetFileMeta(
			*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
			files[idx].FilePath,
			html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
			appHostID,
			metas[idx])
	}
	for idx := range files {
		files[idx].FileMetadata = <-metas[idx]
		if files[idx].FileMetadata.Error != nil {
			ui.Failed("Could not get of file metadata for file %s: %+v", files[idx].FilePath, files[idx].FileMetadata.Error)
			return Failure
		}
	}
	secs := time.Since(start).Seconds()
	log.Tracef("Fetching files metadata took: %.2fs\n", secs)

	// Clean-up HTML5 context
	err = c.CleanHTML5Context(html5Context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	ui.Ok()
	ui.Say("")

	// Display information about HTML5 application files
	table := ui.Table([]string{"path", "size", "etag"})
	for _, file := range files {
		meta := file.FileMetadata
		table.Add(file.FilePath, getReadableSize(meta.FileSize), meta.ETag)
	}
	table.Print()

	return Success
}

// ListSpaceApps get list of applications for current space
func (c *ListCommand) ListSpaceApps() ExecutionStatus {
	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	ui.Say("Getting list of HTML5 applications in org %s / space %s as %s...",
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

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

	// Get list of applications for each app-host service instance
	var data Model
	data.Services = make([]Service, 0)
	for _, serviceInstance := range appHostServiceInstances {
		log.Tracef("Getting list of applications for app-host plan (%+v)\n", serviceInstance)
		applications, err := clients.ListApplicationsForAppHost(*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
			html5Context.HTML5AppRuntimeServiceInstanceKeyToken, serviceInstance.GUID)
		if err != nil {
			ui.Failed("Could not get list of applications for app-host instance %s: %+v", serviceInstance.Name, err)
			return Failure
		}
		apps := make([]App, 0)
		for _, app := range applications {
			apps = append(apps, App{Name: app.ApplicationName, Version: app.ApplicationVersion, Changed: app.ChangedOn, Public: app.IsPublic})
		}
		data.Services = append(data.Services, Service{Name: serviceInstance.Name, GUID: serviceInstance.GUID, UpdatedAt: serviceInstance.UpdatedAt, Apps: apps})
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
	table := ui.Table([]string{"name", "version", "app-host-id", "service instance", "visibility", "last changed"})
	for _, service := range data.Services {
		if len(service.Apps) == 0 {
			table.Add("-", "-", service.GUID, service.Name, "-", service.UpdatedAt)
		} else {
			for _, app := range service.Apps {
				table.Add(app.Name, app.Version, service.GUID, service.Name, (map[bool]string{true: "public", false: "private"})[app.Public], app.Changed)
			}
		}
	}
	table.Print()

	return Success
}

// App app struct
type App struct {
	Name    string
	Version string
	Changed string
	Public  bool
}

// Service service struct
type Service struct {
	UpdatedAt string
	Name      string
	GUID      string
	Apps      []App
	Prefix    string
}

// Model model struct
type Model struct {
	Services []Service
}

// indexOfString returns index of string in array or -1 if not found
func indexOfString(collection []string, value string) int {
	for idx, currentValue := range collection {
		if value == currentValue {
			return idx
		}
	}
	return -1
}

// getReadableSize gets file size in bytes and returns human-readable size in bytes/KB/MB/GB
func getReadableSize(size int) string {
	unit := " bytes"
	if size >= 1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(size)/float64(1024*1024*1024))
	} else if size >= 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/float64(1024*1024))
	} else if size >= 1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/float64(1024))
	}
	return strconv.Itoa(size) + unit
}
