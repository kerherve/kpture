package socket

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/kpture/kpture/pkg/utils"
	"github.com/sirupsen/logrus"
)

type SocketCapture struct {
	Captures []*Capture `json:"captures,omitempty"`
	l        *logrus.Entry
}

func NewSocketCapture(level logrus.Level) *SocketCapture {
	k := SocketCapture{}
	k.l = utils.NewLogger("socket", level)
	return &k
}

func (s *SocketCapture) AddCapture(c *Capture) {
	s.Captures = append(s.Captures, c)
}

func (s *SocketCapture) StartCapture(url string, ch chan []byte) {
	for _, c := range s.Captures {
		StartCapture(c, url, ch)
	}
}

func (s *SocketCapture) MetricServer() {
	s.l.Logger.Info("Setting up Metric server")
	http.HandleFunc("/metrics", s.httpMetric)
	http.ListenAndServe(":8090", nil)
}

func (s *SocketCapture) httpMetric(w http.ResponseWriter, req *http.Request) {
	j, _ := json.Marshal(s.Captures)
	w.Write(j)
}

//Conn is the actual tcp Conn
//Writer is used to write in the pcap file
//capture
func handleConn(Conn net.Conn, Writer *pcapgo.Writer, capture *Capture, ch chan []byte) {

	l := utils.NewLogger("conn", logrus.TraceLevel)
	l = l.WithFields(logrus.Fields{"pod": capture.ContainerName})
	l.Info("Starting capture")
	reader := bufio.NewReader(Conn)

	var buf bytes.Buffer
	buf.Grow(1024)

	for {
		data, err := reader.ReadByte()
		if err == nil {
			if err == io.EOF {
				break
			}
		}
		buf.WriteByte(data)
		if buf.Len() > 16 {
			lenght := int(binary.LittleEndian.Uint32(buf.Bytes()[8:12])) + 16
			if buf.Len() >= lenght {
				packet := buf.Next(lenght)
				Info := gopacket.CaptureInfo{}
				g := binary.LittleEndian.Uint32(packet[0:4])
				n := binary.LittleEndian.Uint32(packet[4:8]) * 1000
				Info.Timestamp = time.Unix(int64(g), int64(n)).UTC()
				Info.CaptureLength = int(binary.LittleEndian.Uint32(packet[8:12]))
				Info.Length = int(binary.LittleEndian.Uint32(packet[12:16]))
				err := Writer.WritePacket(Info, packet[16:])
				//fmt.Println("Sending packet trough channel")
				// p := gopacket.NewPacket(packet[16:], layers.LayerTypeEthernet, gopacket.Default)
				cop := make([]byte, len(packet))
				capture.Stats.NbPacket++
				capture.Stats.NbBytes += uint(Info.Length)
				copy(cop, packet)
				if err != nil {
					l.Error()
				}
				if ch != nil {
					select {
					case ch <- cop:
					default:
						l.Error("could not write in ch")
						l.Info("Channel capacity ", cap(ch))
					}
				}

				// fmt.Println(p.NetworkLayer().NetworkFlow().String())
				// fmt.Println(p)

			}
		}
	}
}

func StartCapture(capture *Capture, url string, ch chan []byte) {
	c, err := net.Dial("tcp", url)
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create(capture.FileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	wf := pcapgo.NewWriter(f)
	wf.WriteFileHeader(1024, layers.LinkTypeEthernet)

	go handleConn(c, wf, capture, ch)

	b, err := json.Marshal(capture.CaptureInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.Write(b)

}
