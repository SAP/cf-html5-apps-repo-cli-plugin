// HTML5 Applications Repository CLI Plugin is a plugin for Cloud Foundry CLI tool
// that aims to provide easy command line access to APIs exposed by HTML5 Application
package main

import (
	"fmt"
	"io"
	defaultlog "log"
	"os"
	"strconv"
	"strings"

	"cf-html5-apps-repo-cli-plugin/commands"
	"cf-html5-apps-repo-cli-plugin/log"

	"cf-html5-apps-repo-cli-plugin/ui"

	"github.com/cloudfoundry/cli/plugin"
)

// Version is the version of the CLI plugin.
var Version = "1.4.9"

// HTML5Plugin represents a cf CLI plugin for working with HTML5 Application Repository service
type HTML5Plugin struct{}

// Commands contains the commands supported by this plugin
var Commands = []commands.Command{
	&commands.ListCommand{},
	&commands.GetCommand{},
	&commands.PushCommand{},
	&commands.DeleteCommand{},
	&commands.InfoCommand{},
}

// Run runs this plugin
func (p *HTML5Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	disableStdOut()
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}
	command, err := findCommand(args[0])
	if err != nil {
		log.Fatalln(err)
	}
	log.Tracef("Running CloudFoundry html5-plugin %s\n", Version)
	err = command.Initialize(command.GetPluginCommand().Name, cliConnection)
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}
	status := command.Execute(args[1:])
	if status == commands.Failure {
		os.Exit(1)
	} else {
		command.Dispose(command.GetPluginCommand().Name)
	}
}

// GetMetadata returns the metadata of this plugin
func (p *HTML5Plugin) GetMetadata() plugin.PluginMetadata {
	metadata := plugin.PluginMetadata{
		Name:          "html5-plugin",
		Version:       parseSemver(Version),
		MinCliVersion: plugin.VersionType{Major: 6, Minor: 7, Build: 0},
	}
	for _, command := range Commands {
		metadata.Commands = append(metadata.Commands, command.GetPluginCommand())
	}
	return metadata
}

func main() {
	plugin.Start(new(HTML5Plugin))
}

func disableStdOut() {
	defaultlog.SetFlags(0)
	defaultlog.SetOutput(io.Discard)
}

func findCommand(name string) (commands.Command, error) {
	for _, command := range Commands {
		pluginCommand := command.GetPluginCommand()
		if pluginCommand.Name == name || pluginCommand.Alias == name {
			return command, nil
		}
	}
	return nil, fmt.Errorf("Could not find command with name '%s'", name)
}

func parseSemver(version string) plugin.VersionType {
	mmb := strings.Split(version, ".")
	if len(mmb) != 3 {
		panic("invalid version: " + version)
	}
	major, _ := strconv.Atoi(mmb[0])
	minor, _ := strconv.Atoi(mmb[1])
	build, _ := strconv.Atoi(mmb[2])

	return plugin.VersionType{
		Major: major,
		Minor: minor,
		Build: build,
	}
}
