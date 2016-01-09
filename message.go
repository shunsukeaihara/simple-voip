package voip

import (
	"errors"
)

type Message interface {
	parse() error
	Process(*VoIPServer) error
}

type JoinMessage struct {
	*Packet
	*Header
	key []byte
}

type PingMessage struct {
	*Packet
	*Header
}

func NewJoinMessage(p *Packet, h *Header) (*JoinMessage, error) {
	m := &JoinMessage{p, h, make([]byte, 0, 0)}
	err := m.parse()
	return m, err
}

func (m *JoinMessage) parse() error {
	if len(m.Data) < 11+int(m.BodySize) {
		return errors.New("body size is too small in join message")
	}
	m.key = m.Data[11 : 11+int(m.BodySize)]
	return nil
}

func (m *JoinMessage) Process(vs *VoIPServer) error {
	// Redis等にkeyを渡して、room_idとuser_idのペアを取得して、roomとconnectionの対応をつける
	uid, rid, err := vs.CheckJoinKey(m.key)
	if err != nil {
		return err
	}
	vs.joinRoom(NewSession(rid, uid, m.Addr))
	return nil
}

func NewPingMessage(p *Packet, h *Header) (*PingMessage, error) {
	m := &PingMessage{p, h}
	return m, nil
}

func (m *PingMessage) parse() error {
	return nil
}

func (m *PingMessage) Process(vs *VoIPServer) error {
	return nil
}
