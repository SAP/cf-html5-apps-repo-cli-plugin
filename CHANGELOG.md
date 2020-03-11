# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Retry for service instance and keys deletion added ([#21](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/21))
- Error handling for service broker errors added ([#16](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/16)) 
- Error handling for `html5-list -a`, when application does not exist ([#19](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/19))
- Support of `--destination` flag for `html5-push`
- Support of `--destination` flag for `html5-list`
- Support of `--destination` flag for `html5-delete`
- Support custom HTML5 runtime for business service destinations with `HTML5_RUNTIME_URL` 
  environment variable (default: `https://<tenant>.cpp.<landscape_url>`)
- Support of `--service` flag for `html5-push`  

### Changed
- List command `html5-list` now checks if first and the only argument passed is an app-host-id,
  and displays list of applications from provided service instace
- Flag `--url` now may be used not only with `--applicaiton` option, but also with `--destination`
  option in `html5-list` command

### Fixed
- Report an error if `html5-delete` failed to delete service instance ([#22](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/22))
- Report an error if cleanup was not successfull ([#23](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/23))
- Don't reuse broken service instances ([#24](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/24))

## [1.3.0] - 2019-08-08
### Added
- Support of `--name` flag for `html5-list` ([#11](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/11))
- Support of `--name` flag for `html5-get` ([#11](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/11))
- Support of `--name` flag for `html5-delete` ([#11](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/11))
- Support of `--name` flag for `html5-info` ([#11](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/11))
- Error handling in `html5-list` and `html5-get` for attempt of accessing private and non-existing applications 

### Fixed
- Use correct syntax for multiple filters in Cloud Foundry API ([#12](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/12))
- Print `1.00 MB` instead of `1024.00 KB` in `html5-info` and `html5-list` ([#13](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/13))
- Use application name and version as `zip` file name, while pushing applications with `html5-push` ([#14](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/13))

## [1.2.0] - 2019-03-01
### Added
- Error handling in `html5-info` for non-existing service instance names and GUIDs
- Rate limit for number of concurrent outgoing connections
- Support of `--name` flag for `html5-push` ([#5](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/5))
- Support of `--redeploy` flag for `html5-push` ([#5](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/5))
- Print detailed error messages in case of client errors in `html5-push` (size exceeded, app already exists)

### Changed
- Change `html5-info` to show service name in first column
- Change `html5-info` to show info about all service instances, if no arguments passed

### Fixed
- Address the case, where `$TMPDIR` is not set on OS ([#4](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/4))
- Fix the problem in `html-info` and `html5-get` with applications containing large amount of files (>1000) ([#7](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/7))

## [1.1.0] - 2019-01-23
### Added
- Support of `--url` flag for `html5-list` command
- Support of `html5-info` command
- Application visibility (private/public) in the list of HTML5 applications
- Support of `html5-delete` command
- Support of multiple app-host-id values in buisiness service binding information
- Metadata, including file size and ETag, is displayed in the list of HTML5 application files
- Support custom HTML5 Application Repository service name via `HTML5_SERVICE_NAME` environment variable (default: `html5-apps-repo`)
- Apache License v2.0 in `LICENSE` file
- Copyright in `NOTICE` file
- Change log in `CANGELOG.md` file

### Changed
- During `html5-push` files are created in system temp folder instead of current working directory ([#2](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/2))

### Fixed
- Normalize paths to files for Windows, replace `\` to `/` in file keys ([#3](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/3))
- Check that directories are apps before trying to upload them with `html5-push` without arguments ([#1](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/1))

## [1.0.0] - 2018-07-26
### Added
- Documentation in `README.md` file
- Support of `html5-list` command
- Support of `html5-get` command
- Support of `html5-push` command

