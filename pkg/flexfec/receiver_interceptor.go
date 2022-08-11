package flexfec

import (
	"os"
	"fmt"
	"sort"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
)

const (
	repairSSRC = 5001
)

type ReceiverInterceptor struct {
	interceptor.NoOp
	recievedBuffer map[Key]rtp.Packet
	repairBuffer   []rtp.Packet
}

func NewReceiverInterceptor() interceptor.Interceptor {
	return &ReceiverInterceptor{
		recievedBuffer: map[Key]rtp.Packet{},
		repairBuffer:   []rtp.Packet{},
	}
}

func (i *ReceiverInterceptor) BindRemoteStream(info *interceptor.StreamInfo, reader interceptor.RTPReader) interceptor.RTPReader {
	fmt.Println("BindRemoteStream called")
	rfile, err := os.Create("output/receiver.txt")

	if err != nil {
		fmt.Println("rfile error", err)
	}

	return interceptor.RTPReaderFunc(func(b []byte, attributes interceptor.Attributes) (int, interceptor.Attributes, error) {
		// recieve packets, and repair packets
		// code
		currPkt := rtp.Packet{}
		currPkt.Unmarshal(b)

		if currPkt.SSRC == repairSSRC {
			// Writing to console
			fmt.Println(string(Blue), "Recieved Repair Packet : ", currPkt.SequenceNumber)

			// Writing to rfile
			fmt.Fprintln(rfile, "Recieved Repair Packet : ")
			fmt.Fprintln(rfile, PrintPkt(currPkt))
			i.repairBuffer = append(i.repairBuffer, currPkt)

		} else {
			Update(i.recievedBuffer, currPkt)

			// Writing to console
			fmt.Println(string(Green), "Received src packet : ", currPkt.SequenceNumber)

			// Writing to rfile
			fmt.Fprintln(rfile, "Received src packet")
			fmt.Fprintln(rfile, PrintPkt(currPkt))
		}

		// recovery phase
		for len(i.repairBuffer) > 0 {
			sort.Slice(i.repairBuffer, func(a, b int) bool {
				return CountMissing(i.recievedBuffer, i.repairBuffer[a]) < CountMissing(i.recievedBuffer, i.repairBuffer[b])
			})

			PrintBuffer(i.repairBuffer)

			//
			currRecPkt := i.repairBuffer[0]
			i.repairBuffer = i.repairBuffer[1:]

			associatedSrcPackets := Extract(i.recievedBuffer, currRecPkt)
			recoveredPacket, status := RecoverMissingPacket(&associatedSrcPackets, currRecPkt)

			if status == 1 {
				fmt.Println(string(White), "Repair packet ", currRecPkt.SequenceNumber, " fully recovered")
			} else if status == 0 {
				fmt.Println(string(White), "Using repair packet ", currRecPkt.SequenceNumber, "to recover")
				fmt.Println("Recovered Packet :", recoveredPacket.SequenceNumber, "\n")

				fmt.Fprintln(rfile, "Recovered packet\n", PrintPkt(recoveredPacket))
				Update(i.recievedBuffer, recoveredPacket)
			} else if status == -1 {
				fmt.Println(string(White), "Recovery not possible")
				fmt.Println()
				i.repairBuffer = append(i.repairBuffer, currRecPkt)
				break
			}

		}

		return reader.Read(b, attributes)
	})
}

// func (i *ReceiverInterceptor) Un
