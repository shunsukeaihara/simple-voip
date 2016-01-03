package voip

import (
	"net"
)

type VoIPServer struct {
	conn *net.UDPConn
}

func NewVoIPServer(addr string) (*VoIPServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	vs := VoIPServer{
		conn: conn,
	}
	return &vs, nil
}
