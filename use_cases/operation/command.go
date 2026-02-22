package operation

// 3. Metadata Command (biar bisa generate HELP otomatis)
type Command struct {
	Name        string
	Description string
	Usage       string
	Type        MsgType
	Payload     []byte
}

type MsgType int

const (
	Shutdown MsgType = iota
	Drain
	Ping
	Pong
	Chat
	EventStats
	ConfigReload
)

// Define the map at package level (optional)
var commandNames = map[MsgType]string{
	Shutdown:     "SHUTDOWN",
	Drain:        "DRAIN",
	Ping:         "PING",
	Chat:         "CHAT",
	Pong:         "PONG",
	EventStats:   "EVENT_STATS",
	ConfigReload: "CONFIG_RELOAD",
}

// String satisfies the fmt.Stringer interface
func (c MsgType) String() string {
	if s, ok := commandNames[c]; ok {
		return s
	}

	return commandNames[Shutdown]
}
