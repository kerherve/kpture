package tcpserver

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/kpture/kpture/pkg/utils"
	log "github.com/sirupsen/logrus"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

type TcpServer struct {
	Logger   *log.Entry
	Receiver chan []byte
	Port     int
	Conns    []net.Conn
}

func NewTcpServer(loglevel log.Level) *TcpServer {
	t := TcpServer{}
	t.Logger = utils.NewLogger("tcpserver", loglevel)
	t.Receiver = make(chan []byte, 100)
	t.Logger.Trace("Setting up Capture Server")

	l, err := net.Listen("tcp4", ":0")
	if err != nil {
		t.Logger.Error(err)
		return nil
	}
	t.Logger.Info("Using port:", l.Addr().(*net.TCPAddr).Port)
	t.Port = l.Addr().(*net.TCPAddr).Port
	go t.setupTcpServer(l)
	go t.handleChannel()

	return &t
}

func (t *TcpServer) setupTcpServer(l net.Listener) {
	t.Conns = []net.Conn{}
	defer l.Close()
	rand.Seed(time.Now().Unix())
	for {
		c, err := l.Accept()
		if err != nil {
			t.Logger.Error(err)
			return
		}
		//When a client connect, we write a pcap file header in it
		wf := pcapgo.NewWriter(c)
		wf.WriteFileHeader(1024, layers.LinkTypeEthernet)
		t.Conns = append(t.Conns, c)
		t.Logger.Info("Handling Capture connect ", c.RemoteAddr().String())
	}
}

func (t *TcpServer) handleChannel() {
	b := []byte{}
	for {
		//When we receive a packet from differents captures routines
		b = <-t.Receiver
		// fmt.Println(b)
		t.Wirte2Conns(b)
	}
}

func (t *TcpServer) Wirte2Conns(b []byte) {
	Info := gopacket.CaptureInfo{}
	g := binary.LittleEndian.Uint32(b[0:4])
	n := binary.LittleEndian.Uint32(b[4:8]) * 1000
	Info.Timestamp = time.Unix(int64(g), int64(n)).UTC()
	Info.CaptureLength = int(binary.LittleEndian.Uint32(b[8:12]))
	Info.Length = int(binary.LittleEndian.Uint32(b[12:16]))

	if Info.CaptureLength != len(b[16:]) {
		t.Logger.Errorf("capture length %d does not match data length %d", Info.CaptureLength, len(b[16:]))
	} else if Info.CaptureLength > Info.Length {
		t.Logger.Errorf("invalid capture info %+v:  capture length > length", Info)
	} else {
		//Go trough all Conns and send the packet
		for i, c := range t.Conns {
			wf := pcapgo.NewWriter(c)
			if err := wf.WritePacket(Info, b[16:]); err != nil {
				t.Logger.Error("Broke pipe, closing conn")
				c.Close()
				t.removeConn(i)
				return
			}
		}

	}
}

func (t *TcpServer) removeConn(i int) {
	t.Conns[i] = t.Conns[0]
	t.Conns = t.Conns[1:]
}
