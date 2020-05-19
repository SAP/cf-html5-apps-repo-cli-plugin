[![GoDoc](https://godoc.org/github.com/SAP/cf-html5-apps-repo-cli-plugin?status.svg)](https://godoc.org/github.com/SAP/cf-html5-apps-repo-cli-plugin)

# CF HTML5 Applications Repository CLI Plugin

[https://sap.github.io/cf-html5-apps-repo-cli-plugin](https://sap.github.io/cf-html5-apps-repo-cli-plugin/)

## Description

The CF HTML5 Applications Repository CLI Plugin is a plugin for the Cloud Foundry CLI tool 
that aims to provide easy command line access to APIs exposed by the HTML5 Application 
Repository service. 
It allows you to:

- Inspect HTML5 applications of current space
- List files of a specific HTML5 application
- View HTML5 applications exposed by business services that are
  bound to the Approuter application
- Download a single file, application or the whole bucket of applications
  uploaded with the same service instance of the `html5-apps-repo` service
- Push one or multiple applications using existing service instances
  of `app-host` plan, or create new ones for you on-the-fly

## Prerequisites

- [Download](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html) and install Cloud Foundry CLI (≥6.36.1).
- [Download](https://golang.org/dl/) and install GO (≥1.11.4).

## Installation

If you want to __use__ the latest released version and don't want to modify it, you
can install the plugin directly from the [Cloud Foundry Community](https://plugins.cloudfoundry.org/#html5-plugin) plugins repository:

```bash
cf install-plugin -r CF-Community "html5-plugin"
```

Otherwise, you can build from source code:
- [Clone or download](https://help.github.com/articles/cloning-a-repository/) the current repository to `/src` folder of your default `GOPATH`:
  * On Unix-like systems: `$HOME/go/src`
  * On Windows systems: `%USERPROFILE%\go\src`
- Open the terminal/console in the root folder of the repository.
- Build sources with `go build`.
- Install the CF HTML5 Applications Repository CLI Plugin:
  * On Unix-like systems: `cf install-plugin -f cf-html5-apps-repo-cli-plugin`
  * On Windows systems: `cf install-plugin -f cf-html5-apps-repo-cli-plugin.exe`

## Upgrade

To upgrade the version of the CF HTML5 Applications Repository CLI Plugin, you need to uninstall the previous version with command:

```bash
cf uninstall-plugin html5-plugin
```

Then install the new version as described in Installation section.

## Usage

The CF HTML5 Applications Repository CLI Plugin supports the following commands:

#### html5-list

<details><summary>History</summary>

| Version  | Changes                                     |
|----------|---------------------------------------------|
| `v1.4.0` | The `--destination` option added            |
| `v1.3.0` | The `--name` option added                   |
| `v1.1.0` | The `--url` option added                    |
| `v1.0.0` | Added in `v1.0.0`                           |

</details>

```
NAME:
   html5-list - Display list of HTML5 applications or file paths of specified application

USAGE:
   cf html5-list [APP_NAME] [APP_VERSION] [APP_HOST_ID|-n APP_HOST_NAME] [-d|-a CF_APP_NAME [-u]]

OPTIONS:
   -APP_NAME          Application name, which file paths should be listed.
                      If not provided, list of applications will be printed.
   -APP_VERSION       Application version, which file paths should be listed.
                      If not provided, the current active version will be used.
   -APP_HOST_ID       GUID of html5-apps-repo app-host service instance that
                      contains application with specified name and version
   -APP_HOST_NAME     Name of html5-apps-repo app-host service instance that 
                      contains the application with specified name and version.
   --name, -n         Use html5-apps-repo app-host service instance name
                      instead of APP_HOST_ID
   --destination, -d  List HTML5 applications exposed via destinations with 
                      sap.cloud.service and html5-apps-repo.app_host_id 
                      properties
   --app, -a          Cloud Foundry application name, which is bound to
                      services that expose UI via html5-apps-repo
   --url, -u          Show conventional URLs of the applications, when accessed 
                      via Cloud Foundry application specified with --app flag
                      or when --destination flag is used                   
```

#### html5-get

<details><summary>History</summary>

| Version  | Changes                                     |
|----------|---------------------------------------------|
| `v1.3.0` | The `--name` option added                   |
| `v1.0.0` | Added in `v1.0.0`                           |

</details>

```
NAME:
   html5-get - Fetch content of single HTML5 application file by path,
               or whole application by name and version

USAGE:
   cf html5-get PATH|APPKEY|--all [APP_HOST_ID|-n APP_HOST_NAME] [--out OUTPUT]

OPTIONS:
   --all              Flag that indicates that all applications of the specified
                      APP_HOST_ID should be fetched
   --out, -o          Output file (for single file) or output directory (for
                      application). By default, standard output and current
                      working directory.
   --name, -n         Use html5-apps-repo app-host service instance name 
                      instead of APP_HOST_ID                   
   -APPKEY            Application name and version
   -APP_HOST_ID       GUID of html5-apps-repo app-host service instance that
                      contains application with specified name and version
   -APP_HOST_NAME     Name of html5-apps-repo app-host service instance that 
                      contains the application with specified name and version                   
   -PATH              Application file path, starting 
                      from /<appName-appVersion>
```

#### html5-push

<details><summary>History</summary>

| Version  | Changes                                           |
|----------|---------------------------------------------------|
| `v1.4.0` | The `--destination` and `--service` options added |
| `v1.2.0` | The `--name` and `--redeploy` options added       |
| `v1.0.0` | Added in `v1.0.0`                                 |

</details>

```
NAME:
   html5-push - Push HTML5 applications to html5-apps-repo service

USAGE:
   cf html5-push [-d|-s SERVICE_INSTANCE_NAME] [-r|-n APP_HOST_NAME] [PATH_TO_APP_FOLDER ...] [APP_HOST_ID]

OPTIONS:
   -APP_HOST_ID              GUID of html5-apps-repo app-host service instance 
                             that contains application with specified name and
                             version
   -APP_HOST_NAME            Name of app-host service instance to which 
                             applications should be deployed
   -PATH_TO_APP_FOLDER       One or multiple paths to folders containing 
                             manifest.json and xs-app.json files
   --destination,-d          Create subaccount level destination with
                             credentials to access HTML5 applications
   --service,-s              Create subaccount level destination with
                             credentials of the service instance
   --name,-n                 Use app-host service instance with specified name
   --redeploy,-r             Redeploy HTML5 applications. All applications
                             should be previously deployed to the same service 
                             instance.
```

#### html5-delete

<details><summary>History</summary>

| Version  | Changes                                     |
|----------|---------------------------------------------|
| `v1.4.0` | The `--destination` option added            |
| `v1.3.0` | The `--name` option added                   |
| `v1.1.0` | Added in `v1.1.0`                           |

</details>

```
NAME:
   html5-delete - Delete one or multiple app-host service instances or content 
                  uploaded with these instances

USAGE:
   cf html5-delete [--content|--destination] APP_HOST_ID|-n APP_HOST_NAME [...]

OPTIONS:
   --content                  delete content only
   --destination,-d           delete destinations that point to service instances to be deleted
   --name,-n                  Use app-host service instance with specified name
   -APP_HOST_ID               GUID of html5-apps-repo app-host service instance
   -APP_HOST_NAME             Name of html5-apps-repo app-host service instance
```

#### html5-info

<details><summary>History</summary>

| Version  | Changes                                     |
|----------|---------------------------------------------|
| `v1.3.0` | The `--name` option added                   |
| `v1.1.0` | Added in `v1.1.0`                           |

</details>

```
NAME:
   html5-info - Get the size limit and status of app-host service instances

USAGE:
   cf html5-info [APP_HOST_ID|-n APP_HOST_NAME ...]

OPTIONS:
   --name,-n          Use app-host service instance with specified name
   -APP_HOST_ID       GUID of html5-apps-repo app-host service instance
   -APP_HOST_NAME     Name of html5-apps-repo app-host service instance
```

## Configuration

The configuration of the CF HTML5 Applications Repository CLI Plugin is done by using environment variables.
The following are supported:
  * `DEBUG=1` - enables trace logs with detailed information about currently running steps
  * `HTML5_CACHE=1` - enables persisted cache. Disabled by default. Should be enabled only for sequential
     execution of the CF HTML5 Applications Repository CLI Plugin commands in the same context 
     (org/space/user) during short period of time (less than 12 hours)
  * `HTML5_SERVICE_NAME` - name of the service in CF marketplace (default: `html5-apps-repo`)
  * `HTML5_RUNTIME_URL` - URL of HTML5 runtime to serve business service 
    destinations (default: `https://<tenant>.cpp.<landscape_url>`)

## Troubleshooting

#### Services and Service Keys

In order to work with HTML5 Application Repository API, the CF HTML5 Applications 
Repository CLI Plugin is required to send JWT with every request. To obtain 
it, the CF HTML5 Applications Repository CLI Plugin creates temporary artifacts, 
such as service instances of `html5-apps-repo` service and service keys for
these service instances. If one of the flows invoked by the CF HTML5 Applications
Repository CLI Plugin fails in the middle, these artifacts may remain
in the current space. 

## Limitations

Currently, you can't use the CF HTML5 Applications Repository CLI Plugin with 
global `-v` flag due to limitations of `cf curl` that is used internally
by the plugin.

## How to obtain support

If you need any support, have any question or have found a bug, please report it in the [GitHub bug tracking system](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues). We shall get back to you.

## License

This project is licensed under the Apache Software License, v. 2 except as noted otherwise in the [LICENSE](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/blob/master/LICENSE) file.

