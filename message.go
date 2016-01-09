package voip

import (
	"bytes"
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
	payload []byte
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
	vs.JoinRoom(NewSession(rid, uid, m.Addr, vs.GetConn()))
	return nil
}

func NewPingMessage(p *Packet, h *Header) (*PingMessage, error) {
	m := &PingMessage{p, h, make([]byte, 0, 0)}
	err := m.parse()
	return m, err
}

func (m *PingMessage) parse() error {
	if len(m.Data) < 11+int(m.BodySize) {
		return errors.New("body size is too small in ping message")
	}
	m.payload = m.Data[11 : 11+int(m.BodySize)]
	return nil
}

func (m *PingMessage) toPong() []byte {
	var b bytes.Buffer
	return b.Bytes()
}

func (m *PingMessage) Process(vs *VoIPServer) error {
	s := vs.GetSession(m.Addr.String())
	if s == nil {
		return errors.New("not authed")
	}
	s.PingChan <- m
	// send pong
	// room := s.GetRoom()
	return nil
}
