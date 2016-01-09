package voip

type Room struct {
	server *VoIPServer
	RoomId int
}

func NewRoom(s *VoIPServer, roomId int) *Room {
	return &Room{s, roomId}
}

func (r *Room) JoinRoom(s *Session) error {
	return nil
}
