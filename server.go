package voip

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type VoIPServer struct {
	conn         *net.UDPConn
	packetChan   chan *Packet
	wg           *sync.WaitGroup
	shutdownChan chan struct{}
	redisCli     *RedisCli
	sessions     map[string]*Session
	sessionM     sync.RWMutex
	rooms        map[int]*Room
	roomM        sync.RWMutex
}

func NewVoIPServer(addr string, numLoop int) (*VoIPServer, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	vs := VoIPServer{
		conn:         conn,
		packetChan:   make(chan *Packet, 100000),
		wg:           new(sync.WaitGroup),
		shutdownChan: make(chan struct{}),
		redisCli:     NewRedisClient("", 0),
		sessions:     make(map[string]*Session),
		sessionM:     sync.RWMutex{},
		rooms:        make(map[int]*Room),
		roomM:        sync.RWMutex{},
	}
	go vs.readLoop()
	for i := 0; i < numLoop; i++ {
		go vs.analyzeLoop()
	}
	return &vs, nil
}

// voipサーバを停止する
func (vs *VoIPServer) Shutdown() {
	close(vs.shutdownChan) //goroutineを止める
	vs.conn.Close()        // socketを止める
	log.Info("waiting all gorutine are stoped")
	vs.wg.Wait()
	// finlize...
}

// UDPソケットからひたすら読み取る
func (vs *VoIPServer) readLoop() {
	var buf [2048]byte
	vs.wg.Add(1)
	defer func() {
		vs.wg.Done()
		vs.Shutdown()
	}()
	for {
		n, addr, err := vs.conn.ReadFromUDP(buf[0:])
		if err != nil {
			// socketが死んでるので落とす -> timeoutかEOS
			return
		}
		vs.packetChan <- &Packet{buf[:n], addr}
	}
}

// パケットをパースして処理する先を決める
func (vs *VoIPServer) analyzeLoop() {
	vs.wg.Add(1)
	defer func() {
		vs.wg.Done()
		vs.Shutdown()
	}()
	for {
		select {
		case p := <-vs.packetChan:
			// packetを解析

			if p == nil {
				continue
			}
			msg, err := p.ToMessage()
			if err == nil {
				// errorの場合は無視
				continue
			}
			msg.Process(vs) // Message毎の処理を実施
		case _, ok := <-vs.shutdownChan:
			if !ok {
				log.Info("shutdown analyze loop")
				return
			}
		}
	}
}

func (vs *VoIPServer) CheckJoinKey(key []byte) (int, int, error) {
	return vs.redisCli.CheckJoinKey(key)
}

// sessionを部屋に結びつける
func (vs *VoIPServer) joinRoom(s *Session) {
	addrStr := s.Addr.String()
	vs.sessionM.RLock()
	_, ok := vs.sessions[addrStr]
	vs.sessionM.RUnlock()
	if ok {
		// already joined
		return
	}
	vs.roomM.RLock()
	room, ok := vs.rooms[s.RoomId]
	vs.roomM.RUnlock()
	if !ok {
		// create room
		room = NewRoom(vs, s.RoomId)
		vs.roomM.Lock()
		vs.rooms[s.RoomId] = room
		vs.roomM.Unlock()
	}
	// join to room
	err := room.JoinRoom(s)
	if err == nil {
		// add to session
		vs.sessionM.Lock()
		vs.sessions[addrStr] = s
		vs.sessionM.Unlock()
	}
}
