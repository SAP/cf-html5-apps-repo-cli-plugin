package commands

import (
	"archive/zip"
	"cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/plugin"
)

// PushCommand fetches the HTML5 application
// file contents
type PushCommand struct {
	HTML5Command
}

// GetPluginCommand returns the plugin command details
func (c *PushCommand) GetPluginCommand() plugin.Command {
	return plugin.Command{
		Name:     "html5-push",
		HelpText: "Push HTML5 applications to html5-apps-repo service",
		UsageDetails: plugin.Usage{
			Usage: "cf html5-push [-r|-n APP_HOST_NAME] [PATH_TO_APP_FOLDER ...] [APP_HOST_ID]",
			Options: map[string]string{
				"-name,-n":           "Use app-host service instance with specified name",
				"-redeploy,-r":       "Redeploy HTML5 applications. All applications should be previously deployed to same service instance",
				"APP_HOST_NAME":      "Name of app-host service instance to which applications should be deployed",
				"PATH_TO_APP_FOLDER": "One or multiple paths to folders containing manifest.json and xs-app.json files",
				"APP_HOST_ID":        "GUID of html5-apps-repo app-host service instance that contains application with specified name and version",
			},
		},
	}
}

// Execute executes plugin command
func (c *PushCommand) Execute(args []string) ExecutionStatus {
	log.Tracef("Executing command '%s': args: '%v'\n", c.Name, args)

	// Parse arguments
	flagSet := flag.NewFlagSet("html5-push", flag.ContinueOnError)
	redeployFlag := flagSet.Bool("redeploy", false, "redeploy HTML5 applications")
	redeployFlagAlias := flagSet.Bool("r", false, "redeploy HTML5 applications")
	nameFlag := flagSet.String("name", "", "app-host service instance name")
	nameFlagAlias := flagSet.String("n", "", "app-host service instance name")
	flagSet.Parse(args)

	// Normalize arguments and aliases
	redeploy := *redeployFlag || *redeployFlagAlias
	log.Tracef("Redeploy flag: %v\n", redeploy)
	serviceName := *nameFlagAlias
	if *nameFlag != "" {
		serviceName = *nameFlag
	}
	log.Tracef("Service name: %v\n", serviceName)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		ui.Failed("Could not get current working directory")
		return Failure
	}

	// No arguments passed
	if len(args) == 0 {
		log.Tracef("No arguments passed. Looking for application directories\n")
		dirs, err := findAppDirectories(cwd)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		return c.PushHTML5Applications(dirs, "", redeploy)
	}

	// Check if passed argument is app-host-id or application
	match := false
	if flagSet.NArg() > 0 {
		log.Tracef("Checking if '%s' is an app-host-id\n", args[len(args)-1])
		match, err = regexp.MatchString("^[A-Za-z0-9]{8}-([A-Za-z0-9]{4}-){3}[A-Za-z0-9]{12}$", args[len(args)-1])
		if err != nil {
			ui.Failed("Regular expression check failed: %+v", err)
			return Failure
		}
	}

	// Validate that app-host-id and app-host name are not passed together
	if match && serviceName != "" {
		ui.Failed("Name of app-host and app-host-id arguments are mutually exclusive. Please use one of them and remove another.")
		return Failure
	}

	// Validate that app-host-id and redeploy are not passed together
	if match && redeploy {
		ui.Failed("Redeploy flag and app-host-id argument are mutually exclusive. Please use one of them and remove another.")
		return Failure
	}

	// Validate that redeploy and app-host name are not passed together
	if redeploy && serviceName != "" {
		ui.Failed("Redeploy flag and app-host name argument are mutually exclusive. Please use one of them and remove another.")
		return Failure
	}

	// Service instance name is passed
	if serviceName != "" {
		// Get context
		log.Tracef("Getting context (org/space/username)\n")
		context, err := c.GetContext()
		if err != nil {
			ui.Failed("Could not get org and space: %s", err.Error())
			return Failure
		}
		// Resolve app-host-id
		log.Tracef("Resolving app-host-id by service instance name '%s'\n", serviceName)
		serviceInstance, err := clients.GetServiceInstanceByName(c.CliConnection, context.SpaceID, serviceName)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		log.Tracef("Resolved app-host-id is '%s'\n", serviceInstance.GUID)
		if flagSet.NArg() == 0 {
			// Only app-host name is provided
			dirs, err := findAppDirectories(cwd)
			if err != nil {
				ui.Failed("%+v", err)
				return Failure
			}
			return c.PushHTML5Applications(dirs, serviceInstance.GUID, redeploy)
		}
		// Both application paths and app-host name are provided
		return c.PushHTML5Applications(flagSet.Args(), serviceInstance.GUID, redeploy)
	}

	// Last argument is app-host-id
	if match {
		log.Tracef("Last argument '%s' is an app-host-id\n", args[len(args)-1])
		// Last argument is app-host-id
		if flagSet.NArg() == 1 {
			// Only app-host-id is provided
			dirs, err := findAppDirectories(cwd)
			if err != nil {
				ui.Failed("%+v", err)
				return Failure
			}
			return c.PushHTML5Applications(dirs, flagSet.Args()[0], redeploy)
		}
		// Both application paths and app-host-id are provided
		return c.PushHTML5Applications(flagSet.Args()[:flagSet.NArg()-1], args[len(args)-1], redeploy)
	}

	// No app directories passed
	if flagSet.NArg() == 0 {
		dirs, err := findAppDirectories(cwd)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		return c.PushHTML5Applications(dirs, "", redeploy)
	}

	// Last argument is application name
	return c.PushHTML5Applications(flagSet.Args(), "", redeploy)
}

// PushHTML5Applications push HTML5 applications to app-host-id
func (c *PushCommand) PushHTML5Applications(appPaths []string, appHostGUID string, redeploy bool) ExecutionStatus {
	var err error
	var zipFiles []string

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	if redeploy || appHostGUID != "" {
		ui.Say("Redeploying HTML5 applications in org %s / space %s as %s...",
			terminal.EntityNameColor(context.Org),
			terminal.EntityNameColor(context.Space),
			terminal.EntityNameColor(context.Username))
	} else {
		ui.Say("Pushing HTML5 applications in org %s / space %s as %s...",
			terminal.EntityNameColor(context.Org),
			terminal.EntityNameColor(context.Space),
			terminal.EntityNameColor(context.Username))
	}

	// Check appPaths are application directories
	dirs := make([]string, 0)
	for _, dir := range appPaths {
		if isAppDirectory(dir) {
			dirs = append(dirs, dir)
		} else {
			ui.Say("%s%s%s",
				terminal.AdvisoryColor("WARNING: Directory '"),
				terminal.EntityNameColor(dir),
				terminal.AdvisoryColor("' is not an application and will not be pushed!\n"))
		}
	}
	if len(dirs) == 0 {
		ui.Failed("Nothing to push. Make sure provided directories contain manifest.json and xs-app.json files")
		return Failure
	}

	// Collect application names
	appNames := make([]string, 0)
	appVersions := make([]string, 0)
	for _, dir := range dirs {
		// Get HTML5 application manifest
		fileName := dir + slash + "manifest.json"
		log.Tracef("Reading %s\n", fileName)
		file, err := os.Open(fileName)
		if err != nil {
			ui.Failed(err.Error())
			return Failure
		}
		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			ui.Failed(err.Error())
			return Failure
		}
		file.Close()

		// Read application name from manifest
		log.Tracef("Extracting application name from: %s\n", string(fileContents))
		var manifest models.HTML5Manifest
		err = json.Unmarshal(fileContents, &manifest)
		if err != nil {
			ui.Failed("Failed to parse manifest.json: %+v", err)
			return Failure
		}
		if manifest.SapApp.ID == "" {
			ui.Failed("Manifest file %s does not define application name (sap.app/id)", fileName)
			return Failure
		}

		// Normalize application name
		appName := strings.Replace(manifest.SapApp.ID, ".", "", -1)
		appName = strings.Replace(appName, "-", "", -1)
		if appName == "" {
			ui.Failed("Manifest file %s defined invalid application name (sap.app/id = '%s')", fileName, manifest.SapApp.ID)
			return Failure
		}
		appNames = append(appNames, appName)

		// Application version
		appVersions = append(appVersions, manifest.SapApp.ApplicationVersion.Version)
	}

	// Find existing app-host
	if appHostGUID == "" && redeploy {

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

		appHostApplicationsMap := make(map[models.CFServiceInstance]models.HTML5ListApplicationsResponse)
		for _, appName := range appNames {
		ServiceInstanceLoop:
			// Look for application name in each app-host service instance
			for _, serviceInstance := range appHostServiceInstances {
				var applications models.HTML5ListApplicationsResponse
				var ok bool
				// Fetch list of app-host applications, if they are not already in cache
				if applications, ok = appHostApplicationsMap[serviceInstance]; !ok {
					log.Tracef("Getting list of applications for app-host plan (%+v)\n", serviceInstance)
					applications, err = clients.ListApplicationsForAppHost(*html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI,
						html5Context.HTML5AppRuntimeServiceInstanceKeyToken, serviceInstance.GUID)
					if err != nil {
						ui.Failed("Could not get list of applications for app-host instance %s: %+v", serviceInstance.Name, err)
						return Failure
					}
					// Store in cache
					appHostApplicationsMap[serviceInstance] = applications
					log.Tracef("List of '%s' service instance applications: %+v\n", serviceInstance.Name, applications)
				}
				for _, app := range applications {
					if app.ApplicationName == appName {
						log.Tracef("Service instance containing application '%s' found (%+v).\n", appName, serviceInstance)
						if appHostGUID != "" && appHostGUID != serviceInstance.GUID {
							ui.Failed("Can't redeploy applications that were previously deployed using different app-host service instances. "+
								"HTML5 application '%s' belongs to app-host '%s' and '%s' belongs to app-host '%s'\n",
								appNames[0], appHostGUID, appName, serviceInstance.GUID)
							return Failure
						}
						appHostGUID = serviceInstance.GUID
						break ServiceInstanceLoop
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

		// Service instance containing application not found
		if appHostGUID == "" {
			ui.Failed("Can't redeploy applications %+v. Applications were not deployed using one of existing service instances", appNames)
			return Failure
		}
	}

	// Create new app-host
	if appHostGUID == "" && !redeploy {

		// Get name of html5-apps-repo service
		serviceName := os.Getenv("HTML5_SERVICE_NAME")
		if serviceName == "" {
			serviceName = "html5-apps-repo"
		}

		// Get context
		log.Tracef("Getting context\n")
		context, err := c.GetContext()
		if err != nil {
			ui.Failed("Could not get context : %+v", err)
			return Failure
		}
		spaceGUID := context.SpaceID

		// Get services
		log.Tracef("Getting list of available services\n")
		services, err := clients.GetServices(c.CliConnection)
		if err != nil {
			ui.Failed("Could not get list of available services : %+v", err)
			return Failure
		}

		// Find html5-apps-repo service
		log.Tracef("Looking for %s service\n", serviceName)
		var serviceGUID string
		for _, service := range services {
			if service.Name == serviceName {
				serviceGUID = service.GUID
			}
		}
		if serviceGUID == "" {
			ui.Failed("Could not find " + serviceName + " service")
			return Failure
		}

		// Get service plan
		log.Tracef("Getting service plans of %s\n", serviceName)
		servicePlans, err := clients.GetServicePlans(c.CliConnection, serviceGUID)
		if err != nil {
			ui.Failed("Could not get service plans for %s : %+v", serviceName, err)
			return Failure
		}

		// Find app-host plan
		log.Tracef("Looking for app-host plan\n")
		var servicePlan *models.CFServicePlan
		for _, plan := range servicePlans {
			if plan.Name == "app-host" {
				servicePlan = &plan
				break
			}
		}
		if servicePlan == nil {
			ui.Failed("Could not find app-host plan of %s service", serviceName)
			return Failure
		}

		// Create service instance
		log.Tracef("Creating service instance for plan %+v\n", *servicePlan)
		serviceInstance, err := clients.CreateServiceInstance(c.CliConnection, spaceGUID, *servicePlan)
		if err != nil {
			ui.Failed("Could not create service instance for %s app-host plan: %+v", serviceName, err)
			return Failure
		}
		appHostGUID = serviceInstance.GUID
	}

	// Create service key for DT
	log.Tracef("Creating service key for app-host-id '%s'\n", appHostGUID)
	serviceKey, err := clients.CreateServiceKey(c.CliConnection, appHostGUID)
	if err != nil {
		ui.Failed("Could not create service key for service instance with id '%s' : %+v", appHostGUID, err)
		return Failure
	}

	// Obtain access token
	log.Tracef("Obtaining access token for service key '%s'\n", serviceKey.Name)
	token, err := clients.GetToken(serviceKey.Credentials)
	if err != nil {
		ui.Failed("Could not obtain access token for service key '': %+v", serviceKey.Name, err)
		return Failure
	}

	// Zip applications
	tmp := os.TempDir()
	if strings.LastIndex(tmp, slash) != len(tmp)-1 {
		tmp = tmp + slash
	}
	zipFiles = make([]string, 0)
	for idx, appPath := range dirs {
		log.Tracef("Zipping the directory: '%s'\n", appPath)

		var appPathFiles = make([]string, 0)
		files, err := ioutil.ReadDir(appPath)
		if err != nil {
			ui.Failed("Could not read contents of application directory '%s' : %+v", appPath, err)
			return Failure
		}
		for _, file := range files {
			log.Tracef("Adding file to archive: '%s'\n", appPath+slash+file.Name())
			appPathFiles = append(appPathFiles, appPath+slash+file.Name())
		}

		zipPath := tmp + appNames[idx] + "-" + appVersions[idx] + ".zip"
		err = zipit(appPathFiles, zipPath)
		if err != nil {
			ui.Failed("Could not zip application directory '%s' : %+v", zipPath, err)
			return Failure
		}
		zipFiles = append(zipFiles, zipPath)
	}

	// Upload zips
	err = clients.UploadAppHost(*serviceKey.Credentials.URI, zipFiles, token)
	if err != nil {
		ui.Failed("Could not upload applications to app-host-id '%s' : %+v", appHostGUID, err)
		return Failure
	}

	// Delete temporarry zip files
	for _, zipFile := range zipFiles {
		log.Tracef("Deleting temporarry zip file: '%s'\n", zipFile)
		err = os.Remove(zipFile)
		if err != nil {
			ui.Failed("Could not delete zip file '%s' : %+v", zipFile, err)
			return Failure
		}
	}

	// Delete temporarry service keys
	log.Tracef("Deleting temporarry service key: '%s'\n", serviceKey.Name)
	err = clients.DeleteServiceKey(c.CliConnection, serviceKey.GUID, maxRetryCount)
	if err != nil {
		ui.Failed("Could not delete service key '%s' : %+v", serviceKey.Name, err)
		return Failure
	}

	ui.Ok()
	ui.Say("")

	return Success
}

func findAppDirectories(cwd string) ([]string, error) {
	// Current working directory
	log.Tracef("Checking if current working directory is an application directory\n")
	if isAppDirectory(cwd) {
		log.Tracef("Pushing current working directory to new app-host-id\n")
		return []string{cwd}, nil
	}
	// Folders in current working directory
	var dirs = make([]string, 0)
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		return dirs, errors.New("Could not read current working directory contents")
	}
	for _, file := range files {
		if file.IsDir() && isAppDirectory(file.Name()) {
			dirs = append(dirs, cwd+slash+file.Name())
		}
	}
	if len(dirs) == 0 {
		return dirs, errors.New("Neither current working directory, nor one of it's subdirectories contains HTML5 application. Make sure manifest.json and xs-app.json exist")
	}
	log.Tracef("Pushing the following directories to new app-host-id: %+v\n", dirs)
	return dirs, nil
}

func isAppDirectory(path string) bool {
	var err error

	// Normalize path
	if string(path[len(path)-1]) != slash {
		path = path + slash
	}

	log.Tracef("Checking if '%s' directory is an application directory\n", path)
	_, err = os.Stat(path + "manifest.json")
	if err != nil {
		return false
	}
	log.Tracef("Directory '%s' contains manifest.json\n", path)
	_, err = os.Stat(path + "xs-app.json")
	if err != nil {
		return false
	}
	log.Tracef("Directory '%s' contains xs-app.json\n", path)

	return true
}

func zipit(sources []string, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return nil
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(source)
		}

		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = strings.Replace(filepath.Join(baseDir, strings.TrimPrefix(path, source)), "\\", "/", -1)
			}

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		})
	}

	return err
}
