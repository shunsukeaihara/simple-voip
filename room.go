package voip

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Room struct {
	server   *VoIPServer
	RoomId   int
	sessions map[int]*Session
	sessionM sync.RWMutex
}

func NewRoom(s *VoIPServer, roomId int) *Room {
	return &Room{
		server:   s,
		RoomId:   roomId,
		sessions: make(map[int]*Session),
		sessionM: sync.RWMutex{},
	}
}

func (r *Room) String() string {
	return fmt.Sprintf("room:%d", r.RoomId)
}

func (r *Room) JoinRoom(s *Session) error {
	r.sessionM.Lock()
	defer r.sessionM.Unlock()
	if _, ok := r.sessions[s.UserId]; !ok {
		r.sessions[s.UserId] = s
		s.SetRoom(r)
		return nil
	} else {
		err := fmt.Errorf("user_id:%d is already joined to room:%d", s.UserId, r.RoomId)
		log.Info(err)
		return err
	}

}
