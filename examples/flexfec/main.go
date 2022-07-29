package main

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/flexfec"
	"github.com/pion/rtp"
)

const (
	listenPort = 6420
	mtu        = 1500
	ssrc       = 5000
)

func sender() {
	serverAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("127.0.0.1:%d", listenPort))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp4", nil, serverAddr)
	if err != nil {
		panic(err)
	}

	sender := flexfec.NewSenderInterceptor(4, 3, false, false)

	streaminfo := interceptor.StreamInfo{
		SSRC: ssrc,
	}

	RTP_writerfunc := interceptor.RTPWriterFunc(func(header *rtp.Header, payload []byte, attributes interceptor.Attributes) (int, error) {
		fmt.Println("Writing to stream")

		currPkt := rtp.Packet{
			Header:  *header,
			Payload: payload,
		}

		fmt.Println(flexfec.PrintPkt(currPkt))

		headerBuf, err := header.Marshal()
		if err != nil {
			panic(err)
		}

		return conn.Write(append(headerBuf, payload...))
	})

	streamWriter := sender.BindLocalStream(&streaminfo, RTP_writerfunc)

	for sequenceNumber := uint16(0); ; sequenceNumber++ {

		packet := flexfec.GenerateRTP(1, 1)[0]
		packet.SequenceNumber = sequenceNumber
		packet.SSRC = ssrc

		// Send a RTP packet with a Payload of 0x0, 0x1, 0x2
		if _, err := streamWriter.Write(&packet.Header, packet.Payload, nil); err != nil {
			fmt.Println(err)
		}

		time.Sleep(time.Millisecond * 200)
	}

}

func receiver() {
	serverAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("127.0.0.1:%d", listenPort))
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", serverAddr)
	if err != nil {
		panic(err)
	}

	receiver := flexfec.NewReceiverInterceptor()

	streaminfo := interceptor.StreamInfo{
		SSRC: ssrc,
	}

	RTP_readerfunc := interceptor.RTPReaderFunc(func(b []byte, _ interceptor.Attributes) (int, interceptor.Attributes, error) {
		return len(b), nil, nil
	})

	streamReader := receiver.BindRemoteStream(&streaminfo, RTP_readerfunc)

	for {
		buffer := make([]byte, mtu)
		i, _, err := conn.ReadFrom(buffer)

		if err != nil {
			panic(err)
		}

		fmt.Println("received RTP")

		if _, _, err := streamReader.Read(buffer[:i], nil); err != nil {
			panic(err)
		}

		currPkt := rtp.Packet{}
		currPkt.Unmarshal(buffer[:i])

		fmt.Println(flexfec.PrintPkt(currPkt))
	}

}

func main() {
	go sender()
	receiver()
}
