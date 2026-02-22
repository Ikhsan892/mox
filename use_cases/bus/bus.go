package bus

import (
	"io"

	"mox/use_cases/operation"
)

type Event struct {
	SourceID string
	Payload  operation.Command
	Output   io.Writer
	Closer   io.Closer
}
