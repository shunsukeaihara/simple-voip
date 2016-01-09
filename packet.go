package voip

import (
	"encoding/binary"
	"errors"
	"net"
)

/*
パケットの構造
1パケットはヘッダ込で500byte以下とする
body[0] message type(uint8)
body[1:10] timestamp(uint64)
body[11:13] bodysize(uint16)
*/

const (
	MSG_TYPE_JOIN = uint8(1)
	MSG_TYPE_PING = uint8(2)
)

type Packet struct {
	Data []byte
	Addr *net.UDPAddr
}

type Header struct {
	MsgType   uint8
	Timestamp uint64
	BodySize  uint16
}

func readHeader(packet []byte) (*Header, error) {
	if len(packet) < 11 {
		// lower than defined header size
		return nil, errors.New("invalid packet size")
	}
	msgType := packet[0]
	timestamp := binary.LittleEndian.Uint64(packet[1:9])
	bodySize := binary.LittleEndian.Uint16(packet[9:11])
	return &Header{msgType, timestamp, bodySize}, nil
}

func (p *Packet) ToMessage() (Message, error) {
	header, err := readHeader(p.Data)
	if err != nil {
		return nil, err
	}
	switch header.MsgType {
	case MSG_TYPE_JOIN:
		return NewJoinMessage(p, header)
	case MSG_TYPE_PING:
		return NewPingMessage(p, header)
	default:
		return nil, errors.New("packet is not a voip message")
	}
}
