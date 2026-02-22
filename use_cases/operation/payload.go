package operation

type MessagePayload struct {
	ID        string
	FromPID   int
	Payload   Command
	Timestamp int64
}
