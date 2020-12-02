package commands

import (
	"cf-html5-apps-repo-cli-plugin/cache"
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

const (
	slash        = string(os.PathSeparator)
	cacheTimeout = 60 * 60
)

var configFilePath = homeDir() + slash +
	".cf" + slash +
	"plugins" + slash +
	"html5-plugin-config.json"

// HTML5Command base struct for HTML5 repository operations
type HTML5Command struct {
	BaseCommand
}

// Initialize initializes the command with the specified name and CLI connection
func (c *HTML5Command) Initialize(name string, cliConnection plugin.CliConnection) {
	log.Tracef("Initializing command '%s'\n", name)
	c.InitializeBase(name, cliConnection)
	if os.Getenv("HTML5_CACHE") == "1" {
		loadCache()
	} else {
		clearCache()
	}
}

// Dispose disposes command and saves cache if needed
func (c *HTML5Command) Dispose(name string) {
	log.Tracef("Disposing command '%s'\n", name)
	c.DisposeBase(name)
	if os.Getenv("HTML5_CACHE") == "1" {
		saveCache()
	}
}

// GetDestinationContext get destination context
func (c *HTML5Command) GetDestinationContext(context Context, destinationInstanceName string) (DestinationContext, error) {

	// Context to return
	var destinationContext = DestinationContext{}

	// Get all services
	log.Tracef("Getting list of services\n")
	services, err := clients.GetServices(c.CliConnection)
	if err != nil {
		return destinationContext, errors.New("Could not get services: " + err.Error())
	}

	// Find destination service
	log.Tracef("Looking for 'destination' service\n")
	var destinationService *models.CFService
	for _, service := range services {
		if service.Name == "destination" {
			destinationService = &service
			break
		}
	}
	if destinationService == nil {
		return destinationContext, fmt.Errorf("Destination service is not in the list of available services." +
			" Make sure your subaccount has entitlement to use it")
	}
	log.Tracef("Destination service found: %+v\n", destinationService)
	destinationContext.DestinationService = destinationService

	// Find destination service "lite" plan
	log.Tracef("Getting service plans for 'destination' service (GUID: %s)\n", destinationService.GUID)
	var liteServicePlan *models.CFServicePlan
	destinationServicePlans, err := clients.GetServicePlans(c.CliConnection, destinationService.GUID)
	if err != nil {
		return destinationContext, fmt.Errorf("Could not get service plans: %s", err.Error())
	}
	for _, servicePlan := range destinationServicePlans {
		if servicePlan.Name == "lite" {
			liteServicePlan = &servicePlan
			break
		}
	}
	if liteServicePlan == nil {
		return destinationContext, fmt.Errorf("Destination service does not have a 'lite' plan")
	}
	log.Tracef("Destination service 'lite' plan found: %+v\n", liteServicePlan)
	destinationContext.DestinationServicePlan = liteServicePlan

	// Get list of service instances of 'lite' plan
	log.Tracef("Getting service instances of 'destination' service 'lite' plan (%+v)\n", liteServicePlan)
	var destinationServiceInstances []models.CFServiceInstance
	destinationServiceInstances, err = clients.GetServiceInstances(c.CliConnection, context.SpaceID, []models.CFServicePlan{*liteServicePlan})
	if err != nil {
		return destinationContext, fmt.Errorf("Could not get service instances for 'lite' plan: %s", err.Error())
	}
	destinationContext.DestinationServiceInstances = destinationServiceInstances

	// Sort destination service instance so that the requested instance to be the first one in the list.
	// If specific destinaton service instance name is required, but not found - return error
	if destinationInstanceName != "" {
		found := false
		for idx, instance := range destinationServiceInstances {
			if instance.Name == destinationInstanceName {
				tmp := destinationServiceInstances[0]
				destinationServiceInstances[0] = instance
				destinationServiceInstances[idx] = tmp
				found = true
				break
			}
		}
		if !found {
			return destinationContext, fmt.Errorf("Could not find service instance of 'destination' service 'lite' plan with name '%s'", destinationInstanceName)
		}
	}

	// Create instance of 'lite' plan if needed
	if len(destinationServiceInstances) == 0 {
		log.Tracef("Creating service instance of 'destination' service 'lite' plan\n")
		destinationServiceInstance, err := clients.CreateServiceInstance(c.CliConnection, context.SpaceID, *liteServicePlan, nil, "")
		if err != nil {
			return destinationContext, fmt.Errorf("Could not create service instance of 'destination' service 'lite' plan: %s", err.Error())
		}
		destinationServiceInstances = append(destinationServiceInstances, *destinationServiceInstance)
		destinationContext.DestinationServiceInstance = destinationServiceInstance
	} else {
		log.Tracef("Using service instance of 'destination' service 'lite' plan: %+v\n", destinationServiceInstances[0])
	}

	// Create service key
	log.Tracef("Creating service key for 'destination' service 'lite' plan\n")
	destinationServiceInstanceKey, err := clients.CreateServiceKey(c.CliConnection, destinationServiceInstances[0].GUID)
	if err != nil {
		return destinationContext, fmt.Errorf("Could not create service key of %s service instance: %s",
			destinationServiceInstances[0].Name,
			err.Error())
	}
	destinationContext.DestinationServiceInstanceKey = destinationServiceInstanceKey

	// Get destination service lite plan key access token
	log.Tracef("Getting token for service key %s\n", destinationServiceInstanceKey.Name)
	destinationServiceInstanceKeyToken, err := clients.GetToken(destinationServiceInstanceKey.Credentials)
	if err != nil {
		return destinationContext, fmt.Errorf("Could not obtain access token: %s", err.Error())
	}
	log.Tracef("Access token for service key %s: %s\n",
		destinationServiceInstanceKey.Name,
		log.Sensitive{Data: destinationServiceInstanceKeyToken})
	destinationContext.DestinationServiceInstanceKeyToken = destinationServiceInstanceKeyToken

	return destinationContext, nil
}

// CleanDestinationContext clean destination context
func (c *HTML5Command) CleanDestinationContext(destinationContext DestinationContext) error {
	var err error

	// Delete service key
	if destinationContext.DestinationServiceInstanceKey != nil {
		log.Tracef("Deleting service key %s\n", destinationContext.DestinationServiceInstanceKey.Name)
		err = clients.DeleteServiceKey(c.CliConnection, destinationContext.DestinationServiceInstanceKey.GUID, maxRetryCount)
		if err != nil {
			return errors.New("Could not delete service key" + destinationContext.DestinationServiceInstanceKey.Name + ": " + err.Error())
		}
		destinationContext.DestinationServiceInstanceKey = nil
	}

	// Delete service instance
	if destinationContext.DestinationServiceInstance != nil {
		log.Tracef("Deleting service instance %s\n", destinationContext.DestinationServiceInstance.Name)
		err = clients.DeleteServiceInstance(c.CliConnection, destinationContext.DestinationServiceInstance.GUID, maxRetryCount)
		if err != nil {
			return errors.New("Could not delete service instance of lite plan: " + err.Error())
		}
		log.Tracef("Service instance %s successfully deleted\n", destinationContext.DestinationServiceInstance.Name)
		destinationContext.DestinationServiceInstance = nil
	}

	return nil
}

// GetHTML5Context get HTML5 context
func (c *HTML5Command) GetHTML5Context(context Context) (HTML5Context, error) {
	log.Tracef("Getting HTML5 context\n")

	// Try to load context from cache
	if html5ContextFromCache, ok := cache.Get("GetHTML5Context:" + context.OrgID + ":" + context.SpaceID); ok {
		log.Tracef("Returning cached HTML5 context\n")
		return html5ContextFromCache.(HTML5Context), nil
	}

	// Context to return
	html5Context := HTML5Context{}

	// Get name of html5-apps-repo service
	serviceName := os.Getenv("HTML5_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "html5-apps-repo"
	}
	html5Context.ServiceName = serviceName

	// Get list of services
	log.Tracef("Getting list of services\n")
	services, err := clients.GetServices(c.CliConnection)
	if err != nil {
		return html5Context, errors.New("Could not get services: " + err.Error())
	}
	html5Context.Services = services

	// Find html5-apps-repo service
	log.Tracef("Looking for '%s' service\n", serviceName)
	var html5AppsRepoService *models.CFService
	for _, service := range services {
		if service.Name == serviceName {
			html5AppsRepoService = &service
			break
		}
	}
	if html5AppsRepoService == nil {
		return html5Context, errors.New(serviceName + " service is not in the list of available services")
	}
	html5Context.HTML5AppsRepoService = html5AppsRepoService

	// Get list of service plans
	log.Tracef("Getting service plans for '%s' service (GUID: %s)\n", serviceName, html5AppsRepoService.GUID)
	servicePlans, err := clients.GetServicePlans(c.CliConnection, html5AppsRepoService.GUID)
	if err != nil {
		return html5Context, errors.New("Could not get service plans: " + err.Error())
	}
	html5Context.HTML5AppsRepoServicePlans = servicePlans

	// Find app-runtime service plan
	log.Tracef("Looking for app-runtime service plan\n")
	var appRuntimeServicePlan *models.CFServicePlan
	for _, plan := range servicePlans {
		if plan.Name == "app-runtime" {
			appRuntimeServicePlan = &plan
			break
		}
	}
	if appRuntimeServicePlan == nil {
		return html5Context, errors.New("Could not find app-runtime service plan")
	}
	html5Context.HTML5AppRuntimeServicePlan = appRuntimeServicePlan

	// Get list of service instances of app-runtime plan
	log.Tracef("Getting service instances of '%s' service app-runtime plan (%+v)\n", serviceName, appRuntimeServicePlan)
	var appRuntimeServiceInstances []models.CFServiceInstance
	appRuntimeServiceInstances, err = clients.GetServiceInstances(c.CliConnection, context.SpaceID, []models.CFServicePlan{*appRuntimeServicePlan})
	if err != nil {
		return html5Context, errors.New("Could not get service instances for app-runtime plan: " + err.Error())
	}

	// Filter out service instances that were recently failed to delete
	for idx, serviceInstance := range appRuntimeServiceInstances {
		if serviceInstance.LastOperation.Type == "delete" && serviceInstance.LastOperation.State == "failed" {
			log.Tracef("Service instance %s is potentially broken and will not be reused\n", serviceInstance.Name)
			appRuntimeServiceInstances[idx] = appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1]
			appRuntimeServiceInstances = appRuntimeServiceInstances[:len(appRuntimeServiceInstances)-1]
		}
	}
	html5Context.HTML5AppRuntimeServiceInstances = appRuntimeServiceInstances

	// Create instance of app-runtime plan if needed
	var appRuntimeServiceInstance *models.CFServiceInstance
	if len(appRuntimeServiceInstances) == 0 {
		log.Tracef("Creating service instance of %s service app-runtime plan\n", serviceName)
		appRuntimeServiceInstance, err = clients.CreateServiceInstance(c.CliConnection, context.SpaceID, *appRuntimeServicePlan, nil, "")
		if err != nil {
			return html5Context, errors.New("Could not create service instance of app-runtime plan: " + err.Error())
		}
		appRuntimeServiceInstances = append(appRuntimeServiceInstances, *appRuntimeServiceInstance)
	}
	html5Context.HTML5AppRuntimeServiceInstance = appRuntimeServiceInstance

	// Get service key
	log.Tracef("Getting list of service keys for service %s\n", appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].Name)
	appRuntimeServiceInstanceKeys, err := clients.GetServiceKeys(c.CliConnection, appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].GUID)
	if err != nil {
		return html5Context, errors.New("Could not get service keys of " +
			appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].Name + " service instance: " + err.Error())
	}
	if len(appRuntimeServiceInstanceKeys) > 0 {
		log.Tracef("Found %d service keys for service %s, using: %+v\n",
			len(appRuntimeServiceInstanceKeys),
			appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].Name,
			appRuntimeServiceInstanceKeys[0])
		html5Context.HTML5AppRuntimeServiceInstanceKey = &appRuntimeServiceInstanceKeys[0]
	}

	// Create service key if needed
	if html5Context.HTML5AppRuntimeServiceInstanceKey == nil {
		log.Tracef("Creating service key for %s service\n", appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].Name)
		appRuntimeServiceInstanceKey, err := clients.CreateServiceKey(c.CliConnection, appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].GUID)
		if err != nil {
			return html5Context, errors.New("Could not create service key of " +
				appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].Name + " service instance: " + err.Error())
		}
		html5Context.HTML5AppRuntimeServiceInstanceKey = appRuntimeServiceInstanceKey
	}

	// Get app-runtime access token
	log.Tracef("Getting token for service key %s\n", html5Context.HTML5AppRuntimeServiceInstanceKey.Name)
	appRuntimeServiceInstanceKeyToken, err := clients.GetToken(html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials)
	if err != nil {
		return html5Context, errors.New("Could not obtain access token: " + err.Error())
	}
	html5Context.HTML5AppRuntimeServiceInstanceKeyToken = appRuntimeServiceInstanceKeyToken
	log.Tracef("Access token for service key %s: %s\n",
		html5Context.HTML5AppRuntimeServiceInstanceKey.Name,
		log.Sensitive{Data: appRuntimeServiceInstanceKeyToken})

	// Runtime URL
	runtimeURL := os.Getenv("HTML5_RUNTIME_URL")
	if runtimeURL == "" {
		uri := *html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.URI
		runtimeURL = "https://" + html5Context.HTML5AppRuntimeServiceInstanceKey.Credentials.UAA.IdentityZone + ".cpp" + uri[strings.Index(uri, "."):]
	}
	html5Context.RuntimeURL = runtimeURL

	// Fill cache
	cache.Set("GetHTML5Context:"+context.OrgID+":"+context.SpaceID, html5Context)

	return html5Context, nil
}

// CleanHTML5Context clean-up temporary service keys and service instances
// created to form HTML5 context
func (c *HTML5Command) CleanHTML5Context(html5Context HTML5Context) error {
	if os.Getenv("HTML5_CACHE") == "1" {
		log.Tracef("Preserving HTML5 context for future use with cache\n")
	} else {
		var err error
		// Delete service key
		if html5Context.HTML5AppRuntimeServiceInstanceKey != nil {
			log.Tracef("Deleting service key %s\n", html5Context.HTML5AppRuntimeServiceInstanceKey.Name)
			err = clients.DeleteServiceKey(c.CliConnection, html5Context.HTML5AppRuntimeServiceInstanceKey.GUID, maxRetryCount)
			if err != nil {
				return errors.New("Could not delete service key" + html5Context.HTML5AppRuntimeServiceInstanceKey.Name + ": " + err.Error())
			}
		}

		// Delete instance of app-runtime if needed
		if html5Context.HTML5AppRuntimeServiceInstance != nil {
			log.Tracef("Deleting service instance %s\n", html5Context.HTML5AppRuntimeServiceInstance.Name)
			err = clients.DeleteServiceInstance(c.CliConnection, html5Context.HTML5AppRuntimeServiceInstance.GUID, maxRetryCount)
			if err != nil {
				return errors.New("Could not delete service instance of app-runtime plan: " + err.Error())
			}
			log.Tracef("Service instance %s successfully deleted\n", html5Context.HTML5AppRuntimeServiceInstance.Name)
		}
	}
	return nil
}

// HTML5Context HTML5 context struct
type HTML5Context struct {
	// Name of html5-apps-repo service in marketplace
	ServiceName string
	// All available CF services
	Services []models.CFService
	// Pointer to html5-apps-repo service
	HTML5AppsRepoService *models.CFService
	// List of html5-apps-repo service plans
	HTML5AppsRepoServicePlans []models.CFServicePlan
	// Pointer to html5-apps-repo app-runtime service plan
	HTML5AppRuntimeServicePlan *models.CFServicePlan
	// Service instances of html5-apps-repo app-runtime service plan
	HTML5AppRuntimeServiceInstances []models.CFServiceInstance
	// Pointer to html5-apps-repo app-runtime service instance
	HTML5AppRuntimeServiceInstance *models.CFServiceInstance
	// Service key of html5-apps-repo app-runtime service plan
	HTML5AppRuntimeServiceInstanceKey *models.CFServiceKey
	// Access token of html5-apps-repo app-runtime service key
	HTML5AppRuntimeServiceInstanceKeyToken string
	// Runtime application URL
	RuntimeURL string
}

// DestinationContext Destination context struct
type DestinationContext struct {
	// Pointer to destination service
	DestinationService *models.CFService
	// Pointer to 'lite' plan of destination service
	DestinationServicePlan *models.CFServicePlan
	// List of destination service instances
	DestinationServiceInstances []models.CFServiceInstance
	// Pointer to destination service instance created during context initialization
	DestinationServiceInstance *models.CFServiceInstance
	// Pointer to destination service key created during context initialization
	DestinationServiceInstanceKey *models.CFServiceKey
	// Access token of destination service key
	DestinationServiceInstanceKeyToken string
}

type stringSlice []string

func (i *stringSlice) String() string {
	return fmt.Sprintf("%d", *i)
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func homeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Could not get user home directory")
	}
	return dir
}

func loadCache() error {
	if _, err := os.Stat(configFilePath); err == nil {
		var config map[string]map[string]json.RawMessage
		var data []byte
		var err error

		// Read configuration file
		data, err = ioutil.ReadFile(configFilePath)
		if err != nil {
			log.Fatalln("Could not read configuration file")
			return err
		}

		// Unmarshal configuration file
		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Fatalln("Could not unmarshal configuration file")
			return err
		}

		// Check timestamp
		if timestamp, ok := config["Timestamp"]; ok {
			if lastUpdated, ok := timestamp["LastUpdated"]; ok {
				var t int64
				err = json.Unmarshal(lastUpdated, &t)
				if err != nil {
					log.Fatalf("Configuration file timestamp last updated has invalid value: %+v", err)
					return err
				}
				age := time.Now().Unix() - t
				if age > cacheTimeout {
					log.Tracef("Configuration file contains outdated cache (%d > %d). Cache will be ignored.\n", age, cacheTimeout)
					delete(config, "Cache")
				} else {
					log.Tracef("Configuration file cache age: %d <= %d (cache timeout)\n", age, cacheTimeout)
				}
			} else {
				log.Fatalln("Configuration file timestamp has invalid structure: no 'LastUpdated' key")
				return err
			}
		} else {
			log.Tracef("Configuration file does not contain timestamp. Cache will be ignored.\n")
			delete(config, "Cache")
		}

		// Lookup for cache
		if cacheConfiguration, ok := config["Cache"]; ok {
			// Load known cache items
			for key, value := range cacheConfiguration {
				if strings.Index(key, "GetHTML5Context:") == 0 {
					var context HTML5Context
					err = json.Unmarshal(value, &context)
					if err != nil {
						log.Fatalln("Could not read HMTL5 context from configuration file cache")
					}
					cache.Set(key, context)
				} else if strings.Index(key, "GetServices:") == 0 {
					var services []models.CFService
					err = json.Unmarshal(value, &services)
					if err != nil {
						log.Fatalln("Could not read HMTL5 services from configuration file cache")
					}
					cache.Set(key, services)
				} else if strings.Index(key, "GetServicePlans:") == 0 {
					var servicePlans []models.CFServicePlan
					err = json.Unmarshal(value, &servicePlans)
					if err != nil {
						log.Fatalln("Could not read HMTL5 service plans from configuration file cache")
					}
					cache.Set(key, servicePlans)
				}
			}
		}

		return nil
	} else if os.IsNotExist(err) {
		log.Traceln("Configuration file not found. Using defaults")
		return err
	} else {
		log.Fatalln("Could not check existence of configuration file")
		return err
	}
}

func saveCache() error {
	if _, err := os.Stat(configFilePath); err == nil {
		var config map[string]interface{}
		var data []byte
		var err error

		log.Traceln("Configuration file found. Updating cache")

		// Read configuration file
		data, err = ioutil.ReadFile(configFilePath)
		if err != nil {
			log.Fatalln("Could not read configuration file")
			return err
		}

		// Unmarshal configuration file
		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Fatalln("Could not unmarshal configuration file")
			return err
		}

		// Update cache
		config["Cache"] = cache.All()
		config["Timestamp"] = map[string]int64{"LastUpdated": time.Now().Unix()}

		// Marshal configuration file
		data, err = json.Marshal(config)
		if err != nil {
			log.Fatalln("Could not marshal configuration file")
			return err
		}

		// Write configuration file
		err = ioutil.WriteFile(configFilePath, data, 0644)
		if err != nil {
			log.Fatalln("Could not write configuration file")
			return err
		}

		return nil
	} else if os.IsNotExist(err) {
		var config map[string]interface{}
		var data []byte
		var err error

		log.Traceln("Configuration file not found. Creating new one")

		// Create configuration
		config = make(map[string]interface{})
		config["Cache"] = cache.All()
		config["Timestamp"] = map[string]int64{"LastUpdated": time.Now().Unix()}

		// Marshal configuration file
		data, err = json.Marshal(config)
		if err != nil {
			log.Fatalln("Could not marshal configuration file")
			return err
		}

		// Write configuration file
		err = ioutil.WriteFile(configFilePath, data, 0644)
		if err != nil {
			log.Fatalln("Could not write configuration file")
			return err
		}

		return nil
	} else {
		log.Fatalln("Could not check existence of configuration file")
		return err
	}
}

func clearCache() error {
	if _, err := os.Stat(configFilePath); err == nil {
		var config map[string]map[string]interface{}
		var data []byte
		var err error

		log.Traceln("Configuration file found. Clearing cache")

		// Read configuration file
		data, err = ioutil.ReadFile(configFilePath)
		if err != nil {
			log.Fatalln("Could not read configuration file")
			return err
		}

		// Unmarshal configuration file
		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Fatalln("Could not unmarshal configuration file")
			return err
		}

		// Update cache
		config["Cache"] = make(map[string]interface{})

		// Marshal configuration file
		data, err = json.Marshal(config)
		if err != nil {
			log.Fatalln("Could not marshal configuration file")
			return err
		}

		// Write configuration file
		err = ioutil.WriteFile(configFilePath, data, 0644)
		if err != nil {
			log.Fatalln("Could not write configuration file")
			return err
		}

		return nil
	} else if os.IsNotExist(err) {
		log.Traceln("Configuration file does not exist. No cache to clear")
		return nil
	} else {
		log.Fatalln("Could not check existence of configuration file")
		return err
	}
}
