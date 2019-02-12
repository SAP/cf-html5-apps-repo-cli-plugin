# HTML5 Applications Repository CLI Plugin

## Description

HTML5 Applications Repository CLI Plugin is a plugin for Cloud Foundry CLI tool 
that aims to provide easy command line access to APIs exposed by HTML5 Application 
Repository service. 
It allows to:

- inspect HTML5 applications of current space
- list files of specific HTML5 application
- view HTML5 applications exposed by business services that are
  bound to Approuter application
- download single file, application or the whole bucket of applications
  uploaded with the same service instance of `html5-apps-repo` service
- push one or multiple applications using existing service instances
  of `app-host` plan, or create new ones for you on-the-fly

## Prerequisites

- [Download](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html) and install Cloud Foundry CLI (≥6.36.1)
- [Download](https://golang.org/dl/) and install GO (≥1.11.4)

## Installation

If you want to __use__ latest released version and not intend to modify it, 
you can install plugin directly from GitHub releses of this repository. For
example, on OSX you can do it as follows (Windows users, add `.exe` at the 
end of the same command):

```bash
cf install-plugin -r CF-Community "html5-plugin"
```

Otherwise, you can build from source code:
- [Clone or download](https://help.github.com/articles/cloning-a-repository/) current repository to `/src` folder of your default `GOPATH`
  * On Unix-like systems `$HOME/go/src`
  * On Windows systems `%USERPROFILE%\go\src`
- Open terminal/console in the root folder of repository
- Build sources with `go build`
- Install CF CLI plugin
  * On Unix-like systems `cf install-plugin -f cf-html5-apps-repo-cli-plugin`
  * On Windows systems `cf install-plugin -f cf-html5-apps-repo-cli-plugin.exe`

## Upgrade

To upgrade version of HTML5 Applications Repository CLI Plugin, you will need to uninstall previous version with command:

```bash
cf uninstall-plugin html5-plugin
```

and then install new version as described in Installation section.

## Usage

The HTML5 Applications Repository CLI Plugin supports the following commands:

#### html5-list

```
NAME:
   html5-list - Display list of HTML5 applications or file paths of specified application

USAGE:
   cf html5-list [APP_NAME] [APP_VERSION] [APP_HOST_ID] [-a CF_APP_NAME [-u]]

OPTIONS:
   -APP_NAME          Application name, which file paths should be listed.
                      If not provided, list of applications will be printed
   -APP_VERSION       Application version, which file paths should be listed.
                      If not provided, current active version will be used
   -APP_HOST_ID       GUID of html5-apps-repo app-host service instance that
                      contains application with specified name and version
   --app, -a          Cloud Foundry application name, which is bound to
                      services that expose UI via html5-apps-repo
   --url, -u          Show conventional URLs of applications, when accessed 
                      via Cloud Foundry application specified with --app flag                   
```

#### html5-get

```
NAME:
   html5-get - Fetch content of single HTML5 application file by path,
               or whole application by name and version

USAGE:
   cf html5-get PATH|APPKEY|--all [APP_HOST_ID] [--out OUTPUT]

OPTIONS:
   --all              Flag that indicates that all applications of specified
                      APP_HOST_ID should be fetched
   --out, -o          Output file (for single file) or output directory (for
                      application). By default, standard output and current
                      working directory
   -APPKEY            Application name and version
   -APP_HOST_ID       GUID of html5-apps-repo app-host service instance that
                      contains application with specified name and version
   -PATH              Application file path, starting 
                      from /<appName-appVersion>
```

#### html5-push

```
NAME:
   html5-push - Push HTML5 applications to html5-apps-repo service

USAGE:
   cf html5-push [PATH_TO_APP_FOLDER ...] [APP_HOST_ID]

OPTIONS:
   -APP_HOST_ID              GUID of html5-apps-repo app-host service instance
                             that contains application with specified name and
                             version
   -PATH_TO_APP_FOLDER       One or multiple paths to folders containing
                             manifest.json and xs-app.json files
```

#### html5-delete (since v1.1.0)

```
NAME:
   html5-delete - Delete one or multiple app-host service instances or content 
                  uploaded with these instances

USAGE:
   cf html5-delete [--content] APP_HOST_ID [...]

OPTIONS:
   --content                  delete content only
   -APP_HOST_ID               GUID of html5-apps-repo app-host service instance
```

#### html5-info (since v1.1.0)

```
NAME:
   html5-info - Get size limit and status of app-host service instances

USAGE:
   cf html5-info APP_HOST_ID [...]

OPTIONS:
   -APP_HOST_ID       GUID of html5-apps-repo app-host service instance
```

## Configuration

The configuration of the HTML5 CLI Plugin is done via environment variables.
The following are supported:
  * `DEBUG=1` - enables trace logs with detailed information about currently running steps
  * `HTML5_SERVICE_NAME` - name of the service in CF marketplace (default: `html5-apps-repo`)

## Troubleshooting

#### Services and Service Keys

In order to work with HTML5 Application Repository API, HTML5 Applications 
Repository CLI Plugin is required to send JWT with every request. To obtain 
it HTML5 Applications Repository CLI Plugin creates temporarry artifacts, 
such as service instances of `html5-apps-repo` service and service keys for
these service instances. If one of the flows invoked by HTML5 Applications
Repository CLI Plugin fails in the middle, these artifacts may remain
in the current space. 

## Limitations

Currently, you can't use HTML5 Applications Repository CLI Plugin with 
global `-v` flag due to limitations of `cf curl` that is used internally
by plugin.

## How to obtain support

If you need any support, have any question or have found a bug, please report it in the [GitHub bug tracking system](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues). We shall get back to you.

## License

This project is licensed under the Apache Software License, v. 2 except as noted otherwise in the [LICENSE](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/blob/master/LICENSE) file.
