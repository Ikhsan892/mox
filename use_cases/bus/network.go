package bus

type NetworkType int

const (
	TCP NetworkType = iota
	UDP
)

var networks = map[NetworkType]string{
	TCP: "tcp",
	UDP: "udp",
}

func (c NetworkType) String() string {
	if s, ok := networks[c]; ok {
		return s
	}

	return networks[TCP]
}
