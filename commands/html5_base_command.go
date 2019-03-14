package commands

import (
	clients "cf-html5-apps-repo-cli-plugin/clients"
	"cf-html5-apps-repo-cli-plugin/clients/models"
	"cf-html5-apps-repo-cli-plugin/log"
	"errors"
	"fmt"
	"os"
)

const (
	slash = string(os.PathSeparator)
)

// HTML5Command base struct for HTML5 repository operations
type HTML5Command struct {
	BaseCommand
}

// GetHTML5Context get HTML5 context
func (c *HTML5Command) GetHTML5Context(context Context) (HTML5Context, error) {

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
	html5Context.HTML5AppRuntimeServiceInstances = appRuntimeServiceInstances

	// Create instance of app-runtime plan if needed
	var appRuntimeServiceInstance *models.CFServiceInstance
	if len(appRuntimeServiceInstances) == 0 {
		log.Tracef("Creating service instance of %s service app-runtime plan\n", serviceName)
		appRuntimeServiceInstance, err = clients.CreateServiceInstance(c.CliConnection, context.SpaceID, *appRuntimeServicePlan)
		if err != nil {
			return html5Context, errors.New("Could not create service instance of app-runtime plan: " + err.Error())
		}
		appRuntimeServiceInstances = append(appRuntimeServiceInstances, *appRuntimeServiceInstance)
	}
	html5Context.HTML5AppRuntimeServiceInstance = appRuntimeServiceInstance

	// Create service key
	log.Tracef("Creating service key for %s service\n", appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].Name)
	appRuntimeServiceInstanceKey, err := clients.CreateServiceKey(c.CliConnection, appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].GUID)
	if err != nil {
		return html5Context, errors.New("Could not create service key of " +
			appRuntimeServiceInstances[len(appRuntimeServiceInstances)-1].Name + " service instance: " + err.Error())
	}
	html5Context.HTML5AppRuntimeServiceInstanceKey = appRuntimeServiceInstanceKey

	// Get app-runtime access token
	log.Tracef("Getting token for service key %s\n", appRuntimeServiceInstanceKey.Name)
	appRuntimeServiceInstanceKeyToken, err := clients.GetToken(appRuntimeServiceInstanceKey.Credentials)
	if err != nil {
		return html5Context, errors.New("Could not obtain access token: " + err.Error())
	}
	html5Context.HTML5AppRuntimeServiceInstanceKeyToken = appRuntimeServiceInstanceKeyToken
	log.Tracef("Access token for service key %s: %s\n", appRuntimeServiceInstanceKey.Name, appRuntimeServiceInstanceKeyToken)

	return html5Context, nil
}

// CleanHTML5Context clean-up temporary service keys and service instances
// created to form HTML5 context
func (c *HTML5Command) CleanHTML5Context(html5Context HTML5Context) error {
	var err error
	// Delete service key
	if html5Context.HTML5AppRuntimeServiceInstanceKey != nil {
		log.Tracef("Deleting service key %s\n", html5Context.HTML5AppRuntimeServiceInstanceKey.Name)
		err = clients.DeleteServiceKey(c.CliConnection, html5Context.HTML5AppRuntimeServiceInstanceKey.GUID)
		if err != nil {
			return errors.New("Could not delete service key" + html5Context.HTML5AppRuntimeServiceInstanceKey.Name + ": " + err.Error())
		}
	}

	// Delete instance of app-runtime if needed
	if html5Context.HTML5AppRuntimeServiceInstance != nil {
		log.Tracef("Deleting service instance %s\n", html5Context.HTML5AppRuntimeServiceInstance.Name)
		err = clients.DeleteServiceInstance(c.CliConnection, html5Context.HTML5AppRuntimeServiceInstance.GUID)
		if err != nil {
			return errors.New("Could not delete service instance of app-runtime plan: " + err.Error())
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
}

type stringSlice []string

func (i *stringSlice) String() string {
	return fmt.Sprintf("%d", *i)
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}
