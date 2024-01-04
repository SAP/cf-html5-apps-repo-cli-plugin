package commands

import (
	"github.com/cloudfoundry/cli/plugin"
)

// Command is an interface that should be implemented by all commands
type Command interface {
	GetPluginCommand() plugin.Command
	Initialize(name string, cliConnection plugin.CliConnection) error
	Execute(args []string) ExecutionStatus
	Dispose(name string)
}
