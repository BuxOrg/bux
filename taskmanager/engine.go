package taskmanager

// Engine is the different types of task manager's that are supported
type Engine string

// Supported engines
const (
	Empty     Engine = "empty"
	Machinery Engine = "machinery"
	TaskQ     Engine = "taskq"
)

// String is the string version of engine
func (e Engine) String() string {
	return string(e)
}

// IsEmpty will return true if the engine is not set
func (e Engine) IsEmpty() bool {
	return e == Empty
}
