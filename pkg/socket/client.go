package socket

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

func handleConn(Conn net.Conn, Writer *pcapgo.Writer, capture Capture, c color.Attribute, MergedFile *pcapgo.Writer) {
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
				b := buf.Bytes()
				packet := buf.Next(lenght)
				Info := gopacket.CaptureInfo{}
				g := binary.LittleEndian.Uint32(packet[0:4])
				n := binary.LittleEndian.Uint32(packet[4:8]) * 1000
				Info.Timestamp = time.Unix(int64(g), int64(n)).UTC()
				Info.CaptureLength = int(binary.LittleEndian.Uint32(packet[8:12]))
				Info.Length = int(binary.LittleEndian.Uint32(packet[12:16]))
				err := Writer.WritePacket(Info, packet[16:])
				err = MergedFile.WritePacket(Info, packet[16:])
				p := gopacket.NewPacket(packet[16:], layers.LayerTypeEthernet, gopacket.Default)
				// fmt.Println(p.NetworkLayer().NetworkFlow().String())
				fmt.Println(p)
				if err != nil {
					fmt.Println(err)
					fmt.Println(string(b))
				}
			}
		}
	}
}

func StartCapture(capture Capture, url string, Writer *pcapgo.Writer) {
	c, err := net.Dial("tcp", url)

	pcolor := rand.Intn(38-30) + 30
	var patr color.Attribute

	switch pcolor {
	case 30:
		patr = color.FgGreen
	case 31:
		patr = color.FgCyan
	case 32:
		patr = color.FgHiYellow
	default:
		patr = color.FgRed
	}

	color.New()
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

	go handleConn(c, wf, capture, patr, Writer)

	b, err := json.Marshal(capture)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.Write(b)

}
