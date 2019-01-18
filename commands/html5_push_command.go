package commands

import (
	"archive/zip"
	"cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"cf-html5-apps-repo-cli-plugin/ui"
	"errors"
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
			Usage: "cf html5-push [PATH_TO_APP_FOLDER ...] [APP_HOST_ID]",
			Options: map[string]string{
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
	var key = "_"
	var argsMap = make(map[string][]string)
	argsMap["_"] = make([]string, 0)
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

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		ui.Failed("Could not get current working directory")
		return Failure
	}

	var argsLength = len(argsMap["_"])
	if argsLength == 0 {
		dirs, err := findAppDirectories(cwd)
		if err != nil {
			ui.Failed("%+v", err)
			return Failure
		}
		return c.PushHTML5Applications(dirs, "")
	}

	// Check if passed argument is app-host-id or application
	log.Tracef("Checking if '%s' is an app-host-id\n", argsMap["_"][argsLength-1])
	match, err := regexp.MatchString("^[A-Za-z0-9]{8}-([A-Za-z0-9]{4}-){3}[A-Za-z0-9]{12}$", argsMap["_"][argsLength-1])
	if err != nil {
		ui.Failed("Regular expression check failed: %+v", err)
		return Failure
	}
	if match {
		log.Tracef("Last argument '%s' is an app-host-id\n", argsMap["_"][argsLength-1])
		// Last argument is app-host-id
		if len(argsMap["_"]) == 1 {
			// Only app-host-id is provided
			dirs, err := findAppDirectories(cwd)
			if err != nil {
				ui.Failed("%+v", err)
				return Failure
			}
			return c.PushHTML5Applications(dirs, argsMap["_"][0])
		}
		// Both application paths and app-host-id are provided
		return c.PushHTML5Applications(argsMap["_"][:argsLength-1], argsMap["_"][argsLength-1])
	}

	// Last argument is application name
	return c.PushHTML5Applications(argsMap["_"], "")
}

// PushHTML5Applications push HTML5 applications to app-host-id
func (c *PushCommand) PushHTML5Applications(appPaths []string, appHostGUID string) ExecutionStatus {
	var err error
	var zipFiles []string

	// Get context
	log.Tracef("Getting context (org/space/username)\n")
	context, err := c.GetContext()
	if err != nil {
		ui.Failed("Could not get org and space: %s", err.Error())
		return Failure
	}

	ui.Say("Pushing HTML5 applications in org %s / space %s as %s...",
		terminal.EntityNameColor(context.Org),
		terminal.EntityNameColor(context.Space),
		terminal.EntityNameColor(context.Username))

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

	// Create new app-host if needed
	if appHostGUID == "" {

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
	zipFiles = make([]string, 0)
	for _, appPath := range dirs {
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

		appPathParts := strings.Split(appPath, slash)
		zipPath := tmp + appPathParts[len(appPathParts)-1] + ".zip"
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
	err = clients.DeleteServiceKey(c.CliConnection, serviceKey.GUID)
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
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if info.IsDir() {
				header.Name += slash
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
