package voip

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	PingInterval = 30
)

type Session struct {
	Addr         *net.UDPAddr
	RoomId       int
	UserId       int
	room         *Room
	conn         *net.UDPConn
	wg           *sync.WaitGroup
	shutdownChan chan struct{}
	PingChan     chan *PingMessage
	mu           sync.RWMutex
}

func NewSession(roomId, userId int, addr *net.UDPAddr, conn *net.UDPConn) *Session {
	s := &Session{
		Addr:         addr,
		RoomId:       roomId,
		UserId:       userId,
		conn:         conn,
		wg:           new(sync.WaitGroup),
		shutdownChan: make(chan struct{}),
		PingChan:     make(chan *PingMessage, 10000),
		mu:           sync.RWMutex{},
	}
	// run gourutines
	go s.pingLoop()
	return s
}

func (s *Session) SetRoom(r *Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.room = r
}

func (s *Session) GetRoom() *Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.room
}

func (s *Session) String() string {
	return fmt.Sprintf("session(addr:%s, room:%d, user:%d)", s.Addr.String(), s.RoomId, s.UserId)
}

func (s *Session) Shutdown() {
	close(s.shutdownChan) //goroutineを止める
	log.Infof("waiting all gorutine in %v are stoped", s)
	s.wg.Wait()
}

func (s *Session) pingLoop() {
	s.wg.Add(1)
	defer func() {
		s.wg.Done()
		s.Shutdown()
	}()
	timer := time.NewTimer(time.Second * time.Duration(PingInterval))
	latestPing := time.Now().Unix()
	for {
		select {
		case _ = <-s.PingChan:
			// update ping recieved time
			latestPing = time.Now().Unix()
			// send pong message
		case <-timer.C:
			if time.Now().Unix()-latestPing >= PingInterval {
				log.Info("ping timeout")
				return
			}
		case _, ok := <-s.shutdownChan:
			if !ok {
				log.Info("shutdown analyze loop")
				return
			}
		}
	}
}
