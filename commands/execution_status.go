package commands

// ExecutionStatus command execution status
type ExecutionStatus int

const (
	// Success command executed sucessfully
	Success ExecutionStatus = 0
	// Failure command failed to execute
	Failure ExecutionStatus = 1
)

// ToInt returns integer representation of command
// execution status
func (status ExecutionStatus) ToInt() int {
	return int(status)
}
