package flexfec

import (
	"fmt"
	"os"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
)

type SenderInterceptor struct {
	interceptor.NoOp
	L           uint8
	D           uint8
	variant     int
	sentPackets map[uint32][]rtp.Packet
}

func NewSenderInterceptor(L, D uint8, variant int) interceptor.Interceptor {
	return &SenderInterceptor{
		L:           L,
		D:           D,
		variant:     variant,
		sentPackets: map[uint32][]rtp.Packet{},
	}
}

func (i *SenderInterceptor) BindLocalStream(info *interceptor.StreamInfo, writer interceptor.RTPWriter) interceptor.RTPWriter {
	fmt.Println("BindLocalStream called")
	file, err := os.Create("output/sender.txt")

	if err != nil {
		fmt.Println("file error", err)
	}

	testcaseMap := GetTestCaseMap(i.variant)

	return interceptor.RTPWriterFunc(func(header *rtp.Header, payload []byte, attributes interceptor.Attributes) (int, error) {
		// fmt.Println(header.SequenceNumber, i.L, i.D)
		_, isPresent := testcaseMap[int(header.SequenceNumber)%int(i.L*i.D)]

		if _, isBufPresent := i.sentPackets[info.SSRC]; !isBufPresent {
			i.sentPackets[info.SSRC] = []rtp.Packet{}
		}

		if len(i.sentPackets[info.SSRC]) < int(i.L*i.D) {
			i.sentPackets[info.SSRC] = append(i.sentPackets[info.SSRC], rtp.Packet{
				Header:  *header,
				Payload: payload,
			})
		} else {

			fmt.Println((White), "Curr src block sender Buffer")
			PrintBuffer(i.sentPackets[info.SSRC]) // printing BUFFER

			currPackets := i.sentPackets[info.SSRC]
			bitstrings := GetBlockBitstring(&currPackets)
			PadBitStrings(&bitstrings, -1)

			// fmt.Println("sorce block starting sn:", i.sentPackets[0].SequenceNumber)

			repairPackets := GenerateRepairLD(&bitstrings, int(i.L), int(i.D), i.variant, i.sentPackets[info.SSRC][0].SequenceNumber)

			fmt.Println(string(Blue), "---sending repiar packets---")
			for _, pkt := range repairPackets {
				// Writing to rfile
				fmt.Fprintln(file, "sending repiar packet")
				fmt.Fprintln(file, PrintPkt(pkt))

				writer.Write(&pkt.Header, pkt.Payload, nil)

			}

			// next block
			i.sentPackets[info.SSRC] = []rtp.Packet{}
			i.sentPackets[info.SSRC] = append(i.sentPackets[info.SSRC], rtp.Packet{
				Header:  *header,
				Payload: payload,
			})
		}

		if isPresent {
			// Writing to console
			fmt.Println(string(Green), "Sending src packet : ", (*header).SequenceNumber)

			// Writing to sender file
			fmt.Fprintln(file, "Sending src packet")
			fmt.Fprintln(file, PrintPkt(rtp.Packet{
				Header:  *header,
				Payload: payload,
			}))

			return writer.Write(header, payload, attributes)
		}

		// Writing to console
		fmt.Println(string(Red), "Missing Src Packet : ", (*header).SequenceNumber)

		// Writing to sender file
		fmt.Fprintln(file, "Missing Src Packet")
		fmt.Fprintln(file, PrintPkt(rtp.Packet{
			Header:  *header,
			Payload: payload,
		}))

		return 0, nil
	})
}

func (i *SenderInterceptor) UnbindLocalStream(info *interceptor.StreamInfo) {
	delete(i.sentPackets, info.SSRC)
}
