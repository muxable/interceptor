package flexfec

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
)

const (
	repairSSRC = 5001
	// Red        = "\033[31m"
	// Green      = "\033[32m"
	// White      = "\033[37m"
	// Blue       = "\033[34m"
)

// const (
// 	repairSSRC = uint32(2868272638)
// 	listenPort = 6420
// 	ssrc       = 5000
// 	mtu        = 200
// )

type ReceiverInterceptor struct {
	interceptor.NoOp
	recievedBuffer map[uint32]map[Key]rtp.Packet
	repairBuffer   []rtp.Packet
	recoveredBuffer map[uint32][]rtp.Packet
}

func NewReceiverInterceptor() interceptor.Interceptor {
	return &ReceiverInterceptor{
		recievedBuffer: map[uint32]map[Key]rtp.Packet{},
		repairBuffer:   []rtp.Packet{},
		recoveredBuffer : map[uint32][]rtp.Packet{},
	}
}

func (i *ReceiverInterceptor) BindRemoteStream(info *interceptor.StreamInfo, reader interceptor.RTPReader) interceptor.RTPReader {
	fmt.Println("BindRemoteStream called")
	file, err := os.Create("output/receiver.txt")

	if err != nil {
		fmt.Println("file error")
	}

	return interceptor.RTPReaderFunc(func(b []byte, attributes interceptor.Attributes) (int, interceptor.Attributes, error) {
		// recieve packets, and repair packets

		

		currPkt := rtp.Packet{}
		currPkt.Unmarshal(b)

		if _, isPresent := i.recievedBuffer[info.SSRC]; !isPresent {
			i.recievedBuffer[info.SSRC] = map[Key]rtp.Packet{}

			// Writing to console
			fmt.Println(string(Green), "Received src packet : ", currPkt.SequenceNumber)
			Update(i.recievedBuffer[info.SSRC], currPkt)

			// Writing to file
			fmt.Fprintln(file, "Recieved src Packet : ")
			fmt.Fprintln(file, PrintPkt(currPkt))

		} else if currPkt.SSRC == repairSSRC {
			// Writing to console
			fmt.Println(string(Blue), "Recieved Repair Packet : ", currPkt.SequenceNumber)

			// Writing to file
			fmt.Fprintln(file, "Recieved Repair Packet : ")
			fmt.Fprintln(file, PrintPkt(currPkt))
			i.repairBuffer = append(i.repairBuffer, currPkt)

		} else {
			// Writing to console
			fmt.Println(string(Green), "Received src packet : ", currPkt.SequenceNumber)
			Update(i.recievedBuffer[info.SSRC], currPkt)

			// Writing to file
			fmt.Fprintln(file, "Recieved src Packet : ")
			fmt.Fprintln(file, PrintPkt(currPkt))
		}

		// recovery phase
		for len(i.repairBuffer) > 0 {
			sort.Slice(i.repairBuffer, func(a, b int) bool {
				return CountMissing(i.recievedBuffer[info.SSRC], i.repairBuffer[a]) < CountMissing(i.recievedBuffer[info.SSRC], i.repairBuffer[b])
			})

			PrintBuffer(i.repairBuffer)

			//
			currRecPkt := i.repairBuffer[0]
			i.repairBuffer = i.repairBuffer[1:]

			associatedSrcPackets := Extract(i.recievedBuffer[info.SSRC], currRecPkt)
			// timing constraint
			recoveryStatusChan := make(chan int, 1)
			recoverPktChan := make(chan rtp.Packet, 1)
			go func() {
				recoveredPacket, status := RecoverMissingPacket(&associatedSrcPackets, currRecPkt)
				recoveryStatusChan <- status
				recoverPktChan <- recoveredPacket

			}()

			var asyncStatus int
			var fetchPacket rtp.Packet = rtp.Packet{}
			select {
			case res := <-recoveryStatusChan:
				// fmt.Println(res)
				asyncStatus = res
				fetchPacket = <-recoverPktChan

				// max timeout of recovery
			case <-time.After(1 * time.Second):
				fmt.Println("timeout")
				fmt.Println("Recovery exited")
			}

			if asyncStatus == 1 {
				fmt.Println(string(White), "Repair packet ", currRecPkt.SequenceNumber, " fully recovered")
			} else if asyncStatus == 0 {
				fmt.Println(string(White), "Using repair packet ", currRecPkt.SequenceNumber, "to recover")
				fmt.Println("Recovered Packet :", fetchPacket.SequenceNumber, "\n")
				
				fmt.Fprintln(file, "Recovered packet\n", PrintPkt(fetchPacket))
				Update(i.recievedBuffer[info.SSRC], fetchPacket)
				i.recoveredBuffer[info.SSRC] = append(i.recoveredBuffer[info.SSRC], fetchPacket)
			} else if asyncStatus == -1 {
				fmt.Println(string(White), "Recovery not possible\n")
				i.repairBuffer = append(i.repairBuffer, currRecPkt)
				break
			} else if asyncStatus == -2 {
				fmt.Println("Either packet to recover was too big or something else is wrong!")
			}

		}

		if _, present := i.recoveredBuffer[info.SSRC]; !present {
			i.recoveredBuffer[info.SSRC] = []rtp.Packet{}
		}

		if len(i.recoveredBuffer[info.SSRC]) > 0 {
			pkt := i.recoveredBuffer[info.SSRC][0]
			i.recoveredBuffer[info.SSRC] = i.recoveredBuffer[info.SSRC][1:]

			buf, err := pkt.Marshal()
			if err != nil {
				fmt.Println("error :", err)
			}

			return copy(b, buf), attributes, nil
		}

		return reader.Read(b, attributes)
	})
}

func (i *ReceiverInterceptor) UnbindRemoteStream(info *interceptor.StreamInfo) {
	delete(i.recievedBuffer, info.SSRC)
}
