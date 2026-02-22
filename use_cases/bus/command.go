package bus

type Command int

const (
	Die Command = iota
	Drain
	Ping
	Pong
)

// Define the map at package level (optional)
var commandNames = map[Command]string{
	Die:   "DIE",
	Drain: "DRAIN",
	Ping:  "PING",
	Pong:  "PONG",
}

// String satisfies the fmt.Stringer interface
func (c Command) String() string {
	if s, ok := commandNames[c]; ok {
		return s
	}

	return "DIE"
}
