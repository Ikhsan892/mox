package exceptions

// 1. Define the enum
type Status int

const (
	Error Status = iota // Zero value for unknown/invalid
	ValidationException
)

var statusNames = map[Status]string{
	Error:               "ERROR",
	ValidationException: "INVALID_INPUT",
}

func (o Status) String() string {
	if name, exists := statusNames[o]; exists {
		return name
	}
	return "unknown"
}
