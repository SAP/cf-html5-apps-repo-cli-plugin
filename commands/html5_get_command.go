package commands

import (
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
)

// GetCommand fetches the HTML5 application
// file contents
type GetCommand struct {
	HTML5Command
}

// GetPluginCommand returns the plugin command details
func (c *GetCommand) GetPluginCommand() plugin.Command {
	return plugin.Command{
		Name:     "html5-get",
		HelpText: "Fetch content of single HTML5 application file by path, or whole application by name and version",
		UsageDetails: plugin.Usage{
			Usage: "cf html5-get PATH|APPKEY|--all [APP_HOST_ID|-n APP_HOST_NAME] [--out OUTPUT]",
			Options: map[string]string{
				"PATH":          "Application file path, starting from /<appName-appVersion>",
				"APPKEY":        "Application name and version",
				"APP_HOST_ID":   "GUID of html5-apps-repo app-host service instance that contains application with specified name and version",
				"APP_HOST_NAME": "Name of html5-apps-repo app-host service instance that contains application with specified name and version",
				"-all":          "Flag that indicates that all applications of specified APP_HOST_ID or APP_HOST_NAME should be fetched",
				"-name, -n":     "Use html5-apps-repo app-host service instance name instead of APP_HOST_ID",
				"-out, -o":      "Output file (for single file) or output directory (for application). By default, standard output and current working directory",
			},
		},
	}
}

// Execute executes plugin command
func (c *GetCommand) Execute(args []string) ExecutionStatus {
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
			ui.Failed("Incorrect number of arguments for APP_HOST_NAME option (expected: 1, actual: %d). For help see [cf html5-get --help]", len(argsMap["--name"]))
			return Failure
		}
		if len(argsMap["--all"]) > 0 || len(argsMap["_"]) == 2 {
			ui.Failed("Can't use both '--name' and APP_HOST_ID at the same time")
			return Failure
		}
		name = argsMap["--name"][0]
	}

	// Output
	var output = ""
	if argsMap["-o"] != nil && argsMap["--out"] != nil {
		ui.Failed("Can't use both '--out' and '-o' at the same time")
		return Failure
	}
	if argsMap["-o"] != nil {
		argsMap["--out"] = argsMap["-o"]
	}
	if argsMap["--out"] != nil {
		if len(argsMap["--out"]) != 1 {
			ui.Failed("Incorrect number of arguments for OUTPUT option (expected: 1, actual: %d). For help see [cf html5-get --help]", len(argsMap["--out"]))
			return Failure
		}
		output = argsMap["--out"][0]
	}

	// Get all apps in app-host by name
	if len(argsMap["--all"]) == 0 && len(argsMap["_"]) == 0 && name != "" {
		return c.GetAppHostFilesContents(output, name, true)
	}

	// Get all apps in app-host-id
	if len(argsMap["--all"]) == 1 {
		return c.GetAppHostFilesContents(output, argsMap["--all"][0], false)
	}

	// Define app-host Name or GUID
	var appHostNameOrGUID = name
	if len(argsMap["_"]) == 2 {
		appHostNameOrGUID = argsMap["_"][1]
	}

	// Get application files or single file
	if len(argsMap["_"]) == 1 || len(argsMap["_"]) == 2 {
		var parts = strings.Split(argsMap["_"][0], "/")
		if len(parts) == 1 {
			// Get application
			var appKeyParts = strings.Split(argsMap["_"][0], "-")
			if len(appKeyParts) == 1 {
				appKeyParts = append(appKeyParts, "")
			}
			return c.GetApplicationFilesContents(output, appKeyParts[0], appKeyParts[1], appHostNameOrGUID, name != "")
		}
		// Get single file
		return c.GetFileContents(output, argsMap["_"][0], appHostNameOrGUID, name != "")
	}

	ui.Failed("Incorrect number of arguments passed. See [cf html5-get --help] for more details")
	return Failure
}

// GetAppHostFilesContents get files contents of all applications of app-host-id
func (c *GetCommand) GetAppHostFilesContents(output string, appHostNameOrGUID string, isName bool) ExecutionStatus {
	log.Tracef("Get content of files of applications of app-host: '%s'\n", appHostNameOrGUID)

	// Channel to control number of concurrent connections
	rateLimiter := make(chan int, maxConcurrentConnections)

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	ui.Say("Getting content of files of applications of app-host %s in org %s / space %s as %s...",
		terminal.EntityNameColor(appHostNameOrGUID),
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

	// Get HTML5 context
	html5Context, err := c.GetHTML5Context(context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	var appHostGUID = appHostNameOrGUID
	if isName {
		// Resolve app-host-id
		log.Tracef("Resolving app-host-id by service instance name '%s'\n", appHostNameOrGUID)
		serviceInstance, err := clients.GetServiceInstanceByName(c.CliConnection, context.SpaceID, appHostNameOrGUID)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		log.Tracef("Resolved app-host-id is '%s'\n", serviceInstance.GUID)
		appHostGUID = serviceInstance.GUID
	}

	var cwd string
	if output == "" {
		// Get current working directory
		cwd, err = os.Getwd()
		if err != nil {
			ui.Failed("Could not get current working directory")
			return Failure
		}
	} else {
		// Use specified directory as output
		cwd = output
	}

	// Normalize (remove trailing slash)
	if string(cwd[len(cwd)-1]) == slash {
		cwd = string(cwd[:len(cwd)-1])
	}

	// Get list of applications for app-host-id
	log.Tracef("Getting list of applications for app-host-id %s\n", appHostGUID)
	applications, err := clients.ListApplicationsForAppHost(*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
		html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
		appHostGUID)
	if err != nil {
		ui.Failed("Could not list of applications for app-host-id %s: %+v", appHostGUID, err)
		return Failure
	}

	var allFiles = make([]models.HTML5ApplicationFile, 0)
	for _, application := range applications {
		var appKey = application.ApplicationName + "-" + application.ApplicationVersion
		// Get list of files for app-host-id and app key
		log.Tracef("Getting list of files for application '%s'\n", appKey)
		files, err := clients.ListFilesOfApp(
			*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
			appKey,
			html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
			appHostGUID)
		if err != nil {
			ui.Failed("Could not list of files for app-host-id '%s': %+v", appHostGUID, err)
			return Failure
		}
		log.Tracef("Number of files for application '%s': %d\n", appKey, len(files))
		allFiles = append(allFiles, files...)
	}

	// Get files
	filesChannels := make([]chan models.HTML5ApplicationFileContent, len(allFiles))
	for idx, file := range allFiles {
		log.Tracef("Getting content of file %s\n", file.FilePath)
		filesChannel := make(chan models.HTML5ApplicationFileContent, 1)
		filesChannels[idx] = filesChannel
		go func(file models.HTML5ApplicationFile, idx int) {
			rateLimiter <- idx
			clients.GetFileContent(
				*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
				file.FilePath,
				html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
				appHostGUID,
				filesChannel)
			log.Tracef("Recieved content of file %s\n", file.FilePath)
		}(file, idx)
	}

	// Save files
	for i := 0; i < len(allFiles); i++ {
		var idx int = <-rateLimiter
		file := allFiles[idx]
		fileContent := <-filesChannels[idx]
		if fileContent.Error != nil {
			ui.Failed("Could not get file contents of %s: %+v", file.FilePath, err)
			return Failure
		}
		// File path
		filePath := cwd + strings.Replace(file.FilePath, "/", slash, -1)
		// Directory path
		dirPath := strings.Split(filePath, slash)
		dirPath = dirPath[:len(dirPath)-1]
		dir := strings.Join(dirPath, slash)
		// Create directory
		log.Tracef("Creating directory %s\n", dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			ui.Failed("Could not create directory %s: %+v", dir, err)
			return Failure
		}
		// Write file
		log.Tracef("Writing file %s\n", filePath)
		err = ioutil.WriteFile(filePath, fileContent.Content, 0644)
		if err != nil {
			ui.Failed("Could not write file %s: %+v", filePath, err)
			return Failure
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

	// Display information about HTML5 application files
	table := ui.Table([]string{"path"})
	for _, file := range allFiles {
		table.Add(file.FilePath)
	}
	table.Print()

	return Success
}

// GetFileContents get file contents
func (c *GetCommand) GetFileContents(output string, filePath string, appHostNameOrGUID string, isName bool) ExecutionStatus {
	log.Tracef("Get content of file with path: '%s'\n", filePath)

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	ui.Say("Getting content of file %s in org %s / space %s as %s...",
		terminal.EntityNameColor(filePath),
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

	// Get HTML5 context
	html5Context, err := c.GetHTML5Context(context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	appHostGUID := appHostNameOrGUID
	if isName {
		// Resolve app-host-id
		log.Tracef("Resolving app-host-id by service instance name '%s'\n", appHostNameOrGUID)
		serviceInstance, err := clients.GetServiceInstanceByName(c.CliConnection, context.SpaceID, appHostNameOrGUID)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		log.Tracef("Resolved app-host-id is '%s'\n", serviceInstance.GUID)
		appHostGUID = serviceInstance.GUID
	}

	// Get file contents
	fileContentChan := make(chan models.HTML5ApplicationFileContent)
	go clients.GetFileContent(
		*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
		filePath,
		html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
		appHostGUID,
		fileContentChan)
	fileContent := <-fileContentChan

	if fileContent.Error != nil {
		ui.Failed("Could not get file contents of %s: %+v", filePath, err)
		return Failure
	}

	// Clean-up HTML5 context
	err = c.CleanHTML5Context(html5Context)
	if err != nil {
		ui.Failed(err.Error())
		return Failure
	}

	if output == "" {
		// Print to stdout
		ui.Ok()
		ui.Say("")
		ui.Say(string(fileContent.Content))
	} else {
		// Directory path
		dirPath := strings.Split(output, slash)
		dirPath = dirPath[:len(dirPath)-1]
		dir := strings.Join(dirPath, slash)
		// Create directory
		log.Tracef("Creating directory %s\n", dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			ui.Failed("Could not create directory %s: %+v", dir, err)
			return Failure
		}
		// Write file
		log.Tracef("Writing file %s\n", output)
		err = ioutil.WriteFile(output, fileContent.Content, 0644)
		if err != nil {
			ui.Failed("Could not write file %s: %+v", output, err)
			return Failure
		}
		ui.Ok()
		ui.Say("")
	}

	return Success
}

// GetApplicationFilesContents get application files contents
func (c *GetCommand) GetApplicationFilesContents(output string, appName string, appVersion string, appHostNameOrGUID string, isName bool) ExecutionStatus {
	log.Tracef("Getting content of application with name: '%s' version: '%s'\n", appName, appVersion)

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

	appHostGUID := appHostNameOrGUID
	if isName {
		// Resolve app-host-id
		log.Tracef("Resolving app-host-id by service instance name '%s'\n", appHostNameOrGUID)
		serviceInstance, err := clients.GetServiceInstanceByName(c.CliConnection, context.SpaceID, appHostNameOrGUID)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		log.Tracef("Resolved app-host-id is '%s'\n", serviceInstance.GUID)
		appHostGUID = serviceInstance.GUID
	}

	// Get list of files
	files, err := clients.ListFilesOfApp(
		*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
		appKey,
		html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
		appHostGUID)
	if err != nil {
		ui.Failed("Could not list of files for app %s: %+v", appKey, err)
		return Failure
	}

	var cwd string
	if output == "" {
		// Get current working directory
		cwd, err = os.Getwd()
		if err != nil {
			ui.Failed("Could not get current working directory")
			return Failure
		}
	} else {
		// Use specified directory as output
		cwd = output
	}

	// Normalize (remove trailing slash)
	if string(cwd[len(cwd)-1]) == slash {
		cwd = string(cwd[:len(cwd)-1])
	}

	// Rate limiter for cuncurrent connections
	rateLimiter := make(chan int, maxConcurrentConnections)

	// Get files
	filesChannels := make([]chan models.HTML5ApplicationFileContent, len(files))
	for idx, file := range files {
		log.Tracef("Getting content of file %s\n", file.FilePath)
		filesChannel := make(chan models.HTML5ApplicationFileContent)
		filesChannels[idx] = filesChannel
		go func(file models.HTML5ApplicationFile, idx int) {
			rateLimiter <- idx
			clients.GetFileContent(
				*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
				file.FilePath,
				html5Context.HTML5AppRuntimeServiceInstanceKeyToken,
				appHostGUID,
				filesChannel)
		}(file, idx)
	}

	// Save files
	for i := 0; i < len(files); i++ {
		var idx int = <-rateLimiter
		file := files[idx]
		fileContent := <-filesChannels[idx]
		if fileContent.Error != nil {
			ui.Failed("Could not get file contents of %s: %+v", file.FilePath, err)
			return Failure
		}
		// File path
		filePath := cwd + strings.Replace(file.FilePath, "/", slash, -1)
		// Directory path
		dirPath := strings.Split(filePath, slash)
		dirPath = dirPath[:len(dirPath)-1]
		dir := strings.Join(dirPath, slash)
		// Create directory
		log.Tracef("Creating directory %s\n", dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			ui.Failed("Could not create directory %s: %+v", dir, err)
			return Failure
		}
		// Write file
		log.Tracef("Writing file %s\n", filePath)
		err = ioutil.WriteFile(filePath, fileContent.Content, 0644)
		if err != nil {
			ui.Failed("Could not write file %s: %+v", filePath, err)
			return Failure
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

	// Display information about HTML5 application files
	table := ui.Table([]string{"path"})
	for _, file := range files {
		table.Add(file.FilePath)
	}
	table.Print()

	return Success
}
