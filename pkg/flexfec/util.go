package flexfec

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pion/rtp"
)

const (
	Red        = "\033[31m"
	Green      = "\033[32m"
	White      = "\033[37m"
	Blue       = "\033[34m"
)

func payloadGenerator(size int) []byte {

	payload := make([]byte, size)

	for i := 0; i < size; i++ {
		// rand.Seed(time.Now().UnixNano())
		randContent := rand.Intn(256)
		payload[i] = byte(randContent)
	}

	return payload
}

// Creates L * D RTP packets with variable payload size
func GenerateRTP(L int, D int) []rtp.Packet {
	rand.Seed(time.Now().UnixNano())

	// pkts := []byte{
	// 	0x90, 0xE0, 0x69, 0x8F, 0xD9, 0xC2, 0x93, 0xDA,
	// 	0x1C, 0x64, 0x27, 0x82, 0x45, 0xB1, 0x34, 0xF1,
	// 	0xFF, 0xFF, 0xAF, 0xFF, 0x98, 0x36, 0xbE, 0x88,
	// }

	csrc := []uint32{
		0x64, 0x27, 0x82, 0x45, 0x90, 0xE0,
	}

	// size := len(pkts)

	n := uint16(L * D)
	// SN_base := uint16(rand.Intn(65535 - int(n)))
	SN_base := uint16(10000)
	ssrc := uint32(rand.Intn(4294967296))

	packets := []rtp.Packet{}

	for i := uint16(0); i < n; i++ {
		endIndex := 10 + rand.Intn(14)
		csrcIndex := rand.Intn(5)
		headerExtensionIndex := rand.Intn(3)
		isExtension := rand.Intn(2)

		packet := rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				Padding:        false,
				Extension:      false,
				Marker:         false,
				PayloadType:    15,
				SequenceNumber: SN_base + i,
				Timestamp:      54243243,
				SSRC:           ssrc,
				CSRC:           csrc[:csrcIndex],
			},
			Payload: payloadGenerator(endIndex),
		}

		if isExtension == 1 {
			packet.Header.Extension = true
			packet.Header.ExtensionProfile = 0x1000
			packet.Header.SetExtension(uint8(1), payloadGenerator(headerExtensionIndex))
			packet.Header.SetExtension(uint8(2), payloadGenerator(headerExtensionIndex))
			packet.Header.SetExtension(uint8(3), payloadGenerator(headerExtensionIndex))
		}
		packets = append(packets, packet)
	}

	return packets
}

func PrintPkt(pkt rtp.Packet) string {
	result := fmt.Sprintf("Header : %v\n", pkt.Header)
	result += fmt.Sprintf("Payload : %v\n", pkt.Payload)
	result += fmt.Sprintf("PaddingSize : %d\n\n", pkt.PaddingSize)
	return result
}

func PrintBytes(buf []byte) {
	for index, value := range buf {
		for i := 7; i >= 0; i-- {
			fmt.Print((value >> i) & 1)
		}
		fmt.Print(" ")
		if (index+1)%4 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func PrintBuffer(Buffer []rtp.Packet) {
	fmt.Print("BUFFER : [ ")
	for _, pkt := range Buffer {
		fmt.Print(pkt.SequenceNumber, " ")
	}
	fmt.Println("]")
}

func GetTestCaseMap(variant int) map[int]int {

	if variant == 0 {
		/*
			0 1  2  X
			4 5  X  7
			X 9 10 11
			c1 c2 c3 c4
		*/

		return map[int]int{
			0: 1,
			1: 1,
			2: 1,
			// 3 : 1,
			4: 1,
			5: 1,
			6: 1,
			7: 1,
			// 8 : 1,
			9:  1,
			10: 1,
			11: 1,
		}
	} else if variant == 1 {
		/*
			0 X  2  X
			4 5  6  7
			X 9  X 11
			c1 c2 c3 c4
		*/

		return map[int]int{
			0: 1,
			// 1 : 1,
			2: 1,
			// 3 : 1,
			4: 1,
			5: 1,
			6: 1,
			7: 1,
			// 8 : 1,
			9: 1,
			// 10 : 1,
			11: 1,
		}
	}

	/*
		X X  2  3 |0
		4 X  X  7 |1
		8 9  X  X |2
		3 4  5  6
	*/

	return map[int]int{
		// 0 : 1,
		// 1 : 1,
		2: 1,
		3: 1,
		4: 1,
		// 5 : 1,
		// 6 : 1,
		7: 1,
		8: 1,
		9: 1,
		// 10 : 1,
		// 11 : 1,
	}

}
