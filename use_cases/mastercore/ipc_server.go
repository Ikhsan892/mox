package mastercore

import (
	"net"

	"mox/use_cases/workercore"
)

type IPCServer struct {
	socketPath string
	port       int
	registrar  workercore.IRegistrar
}

func (c *IPCServer) handleHandshake(conn *net.UnixConn) {}

func (c *IPCServer) sendFD(conn *net.UnixConn, fd int) {}
