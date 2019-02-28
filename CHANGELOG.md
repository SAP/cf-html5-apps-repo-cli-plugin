# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

