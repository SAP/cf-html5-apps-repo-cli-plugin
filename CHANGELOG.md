# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
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
- Check that directories are apps before trying to upload them with `html5-push` without arguments ([#1](https://github.com/SAP/cf-html5-apps-repo-cli-plugin/issues/1))

## [1.0.0] - 2018-07-26
### Added
- Documentation in `README.md` file
- Support of `html5-list` command
- Support of `html5-get` command
- Support of `html5-push` command

