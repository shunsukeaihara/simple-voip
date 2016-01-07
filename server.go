package voip

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type VoIPServer struct {
	conn         *net.UDPConn
	packetChan   chan *Packet
	mu           *sync.RWMutex
	wg           *sync.WaitGroup
	shutdownChan chan struct{}
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
		packetChan:   make(chan *Packet, 10000),
		mu:           new(sync.RWMutex),
		wg:           new(sync.WaitGroup),
		shutdownChan: make(chan struct{}),
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
			msg.Process() // Message毎の処理を実施
		case _, ok := <-vs.shutdownChan:
			if !ok {
				log.Info("shutdown analyze loop")
				return
			}
		}
	}
}
