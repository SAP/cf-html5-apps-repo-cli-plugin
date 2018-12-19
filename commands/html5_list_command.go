package commands

import (
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"strconv"

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
			Usage: "cf html5-list [APP_NAME] [APP_VERSION] [APP_HOST_ID] [-a CF_APP_NAME]",
			Options: map[string]string{
				"APP_NAME":    "Application name, which file paths should be listed. If not provided, list of applications will be printed",
				"APP_VERSION": "Application version, which file paths should be listed. If not provided, current active version will be used",
				"APP_HOST_ID": "GUID of html5-apps-repo app-host service instance that contains application with specified name and version",
				"-app, -a":    "Cloud Foundry application name, which is bound to services that expose UI via html5-apps-repo",
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
			continue
		}
		if argsMap[key] == nil {
			argsMap[key] = make([]string, 0)
		}
		argsMap[key] = append(argsMap[key], arg)
		key = "_"
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

	if app != "" {
		// List HTML5 applications available in CF application context
		return c.ListAppApps(app)
	} else if len(argsMap["_"]) == 3 {
		// List files paths of application with version from app-host-id
		return c.ListAppFiles(argsMap["_"][0], argsMap["_"][1], argsMap["_"][2])
	} else if len(argsMap["_"]) == 2 {
		// List files paths of application with version
		return c.ListAppFiles(argsMap["_"][0], argsMap["_"][1], "")
	} else if len(argsMap["_"]) == 1 {
		// List files paths of application default version
		return c.ListAppFiles(argsMap["_"][0], "", "")
	}

	ui.Failed("Too much arguments. See [cf html5-list --help] for more detals")
	return Failure
}

// ListAppApps get list of HTML5 applications available in CF application context
func (c *ListCommand) ListAppApps(appName string) ExecutionStatus {
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
			apps = append(apps, App{Name: app.ApplicationName, Version: app.ApplicationVersion, Changed: app.ChangedOn})
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
				// Get list of applications for app-host-id
				log.Tracef("Getting list of applications for service '%s' and app-host-id '%s'\n", serviceName, serviceBinding.Credentials.HTML5AppsRepo.AppHostID)
				applications, err := clients.ListApplicationsForAppHost(*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
					html5Context.HTML5AppRuntimeServiceInstanceKeyToken, serviceBinding.Credentials.HTML5AppsRepo.AppHostID)
				if err != nil {
					ui.Failed("Could not get list of applications for app-host-id '%s': %+v", serviceBinding.Credentials.HTML5AppsRepo.AppHostID, err)
					return Failure
				}
				apps := make([]App, 0)
				for _, app := range applications {
					apps = append(apps, App{Name: app.ApplicationName, Version: app.ApplicationVersion, Changed: app.ChangedOn})
				}
				servicesData.Services = append(servicesData.Services, Service{GUID: serviceBinding.Credentials.HTML5AppsRepo.AppHostID, Name: serviceName, Apps: apps})
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

	// Display information about HTML5 applications
	table := ui.Table([]string{"name", "version", "app-host-id", "service instance", "last changed"})
	for _, service := range data.Services {
		for _, app := range service.Apps {
			table.Add(app.Name,
				terminal.LogStdoutColor(app.Version),
				terminal.LogStdoutColor(service.GUID),
				terminal.LogStdoutColor(service.Name),
				terminal.LogStdoutColor(app.Changed))
		}
	}
	for _, service := range servicesData.Services {
		for _, app := range service.Apps {
			table.Add(app.Name,
				terminal.AdvisoryColor(app.Version),
				terminal.AdvisoryColor(service.GUID),
				terminal.AdvisoryColor(service.Name),
				terminal.AdvisoryColor(app.Changed))
		}
	}
	table.Print()

	return Success
}

// ListAppFiles get list of application files
func (c *ListCommand) ListAppFiles(appName string, appVersion string, appHostID string) ExecutionStatus {
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
	for idx := range files {
		meta, err := clients.GetFileMeta(
			*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
			files[idx].FilePath,
			html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
			appHostID)
		if err != nil {
			ui.Failed("Could not get of file metadata for file %s: %+v", files[idx].FilePath, err)
			return Failure
		}
		files[idx].FileMetadata = meta
	}

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
			apps = append(apps, App{Name: app.ApplicationName, Version: app.ApplicationVersion, Changed: app.ChangedOn})
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
	table := ui.Table([]string{"name", "version", "app-host-id", "service instance", "last changed"})
	for _, service := range data.Services {
		if len(service.Apps) == 0 {
			table.Add("-", "-", service.GUID, service.Name, service.UpdatedAt)
		} else {
			for _, app := range service.Apps {
				table.Add(app.Name, app.Version, service.GUID, service.Name, app.Changed)
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
}

// Service service struct
type Service struct {
	UpdatedAt string
	Name      string
	GUID      string
	Apps      []App
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
	if size > 1024*1024*1024 {
		return strconv.Itoa(size/(1024*1024*1024)) + " GB"
	} else if size > 1024*1024 {
		return strconv.Itoa(size/(1024*1024)) + " MB"
	} else if size > 1024 {
		return strconv.Itoa(size/1024) + " KB"
	}
	return strconv.Itoa(size) + unit
}
