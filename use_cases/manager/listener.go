package manager

import "net"

type ListenerManager struct {
	ports []int
	conns map[int]net.Listener
}
