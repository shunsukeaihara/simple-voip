package voip

import ()

type Message interface {
	Parse() error
	Process() error
}

type JoinMessage struct {
	*Packet
	*Header
	vs *VoIPServer
}

type PingMessage struct {
	*Packet
	*Header
	vs *VoIPServer
}

func NewJoinMessage(p *Packet, h *Header) *JoinMessage {
	return &JoinMessage{p, h, nil}
}

func (m *JoinMessage) Parse() error {
	return nil
}

func (m *JoinMessage) Process() error {
	return nil
}

func NewPingMessage(p *Packet, h *Header) *PingMessage {
	return &PingMessage{p, h, nil}
}

func (m *PingMessage) Parse() error {
	return nil
}

func (m *PingMessage) Process() error {
	return nil
}
