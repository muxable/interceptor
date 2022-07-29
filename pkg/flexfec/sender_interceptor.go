package flexfec

import (
	"fmt"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
)

type SenderInterceptor struct {
	interceptor.NoOp
	L           uint8
	D           uint8
	r           bool
	f           bool
	sentPackets []rtp.Packet
}

func NewSenderInterceptor(L, D uint8, r, f bool) interceptor.Interceptor {
	return &SenderInterceptor{
		L: L,
		D: D,
		r: r,
		f: f,
	}
}

func (i *SenderInterceptor) BindLocalStream(info *interceptor.StreamInfo, writer interceptor.RTPWriter) interceptor.RTPWriter {
	fmt.Println("BindLocalStream called")

	return interceptor.RTPWriterFunc(func(header *rtp.Header, payload []byte, attributes interceptor.Attributes) (int, error) {
		if len(i.sentPackets) < int(i.L*i.D) {
			i.sentPackets = append(i.sentPackets, rtp.Packet{
				Header:  *header,
				Payload: payload,
			})
		} else if !i.r && !i.f { // rowFEC

			fmt.Println("Curr src block")
			fmt.Println(i.sentPackets)

			// row fec
			bitsrings := GetBlockBitstring(&i.sentPackets)
			PadBitStrings(&bitsrings, -1)

			fmt.Println("generate row repair for source block")
			fmt.Println("sorce block starting sn:", i.sentPackets[0].SequenceNumber)
			rowRepairpackets := GenerateRepairRowFec(&bitsrings, int(i.L), false, i.sentPackets[0].SequenceNumber)

			fmt.Println("sending row repair")
			for _, pkt := range rowRepairpackets {
				writer.Write(&pkt.Header, pkt.Payload, nil)
			}

			// next block
			i.sentPackets = []rtp.Packet{}
			i.sentPackets = append(i.sentPackets, rtp.Packet{
				Header:  *header,
				Payload: payload,
			})
		}

		return writer.Write(header, payload, attributes)
	})
}
