package log

import (
	"fmt"
	"os"
	"time"
)

// Debug true if environment variable DEBUG is set to 1 or 2
var Debug = (os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "2")

// DebugSensitive true if environment variable DEBUG is set to 2
var DebugSensitive = (os.Getenv("DEBUG") == "2")

// Sensitive is a wrapper around any sensitive data
// that should not be logged, except the corresponding
// flag is set to true
type Sensitive struct {
	Data interface{}
}

func (s Sensitive) String() string {
	if DebugSensitive {
		return fmt.Sprintf("+%v", s.Data)
	}
	return "[ SENSITIVE DATA ]"
}

// Exiter exiter interface
type Exiter interface {
	Exit(status int)
}

// DefaultExiter default exiter structure
type DefaultExiter struct {
}

// Exit exit program with status
func (e DefaultExiter) Exit(status int) {
	os.Exit(status)
}

var exiter Exiter = DefaultExiter{}

// GetExiter returns exiter
func GetExiter() Exiter {
	return exiter
}

// SetExiter sets exiter
func SetExiter(e Exiter) {
	exiter = e
}

// Exit calls exiter.Exit with status
func Exit(status int) {
	exiter.Exit(status)
}

// Fatal prints value and exits with status 1
func Fatal(v ...interface{}) {
	Print(v...)
	exiter.Exit(1)
}

// Fatalf prints formatted value and exits with status 1
func Fatalf(format string, v ...interface{}) {
	Printf(format, v...)
	exiter.Exit(1)
}

// Fatalln prints value line and exits with status 1
func Fatalln(v ...interface{}) {
	Println(v...)
	exiter.Exit(1)
}

// Print prints value
func Print(v ...interface{}) {
	fmt.Print(v...)
}

// Printf prints formatted value
func Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

// Println print value line
func Println(v ...interface{}) {
	fmt.Println(v...)
}

// Trace print value if Debug flag is on
func Trace(v ...interface{}) {
	if Debug {
		Print(time.Now().Format("15:04:05.000000") + " ")
		Print(v...)
	}
}

// Tracef print formatted value if Debug flag is on
func Tracef(format string, v ...interface{}) {
	if Debug {
		Print(time.Now().Format("15:04:05.000000") + " ")
		Printf(format, v...)
	}
}

// Traceln print value line if Debug flag is on
func Traceln(v ...interface{}) {
	if Debug {
		Print(time.Now().Format("15:04:05.000000") + " ")
		Println(v...)
	}
}
