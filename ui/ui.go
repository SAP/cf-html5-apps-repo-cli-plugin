package ui

import (
	"os"

	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
)

var teePrinter *terminal.TeePrinter
var ui terminal.UI

func init() {
	i18n.T = func(translationID string, args ...interface{}) string {
		return translationID
	}
	teePrinter = terminal.NewTeePrinter()
	ui = terminal.NewUI(os.Stdin, teePrinter)
	teePrinter.DisableTerminalOutput(false)
}

// SetOutputBucket set output bucket
func SetOutputBucket(bucket *[]string) {
	teePrinter.SetOutputBucket(bucket)
}

// ClearOutputBucket clear output bucket
func ClearOutputBucket() {
	teePrinter.SetOutputBucket(nil)
}

// DisableTerminalOutput disable terminal output
func DisableTerminalOutput(disable bool) {
	teePrinter.DisableTerminalOutput(disable)
}

// PrintPaginator print paginator
func PrintPaginator(rows []string, err error) {
	ui.PrintPaginator(rows, err)
}

// Say say
func Say(message string, args ...interface{}) {
	ui.Say(message, args...)
}

// PrintCapturingNoOutput print capturing no output
func PrintCapturingNoOutput(message string, args ...interface{}) {
	ui.PrintCapturingNoOutput(message, args...)
}

// Warn warning
func Warn(message string, args ...interface{}) {
	ui.Warn(message, args...)
}

// Ask ask
func Ask(prompt string, args ...interface{}) (answer string) {
	return ui.Ask(prompt, args...)
}

// Confirm confirm
func Confirm(message string, args ...interface{}) bool {
	return ui.Confirm(message, args...)
}

// Ok ok
func Ok() {
	ui.Ok()
}

// Failed failed
func Failed(message string, args ...interface{}) {
	ui.Failed(message, args...)
}

// PanicQuietly panic quietly
func PanicQuietly() {
	ui.PanicQuietly()
}

// LoadingIndication loading indication
func LoadingIndication() {
	ui.LoadingIndication()
}

// Table table
func Table(headers []string) terminal.Table {
	return ui.Table(headers)
}
