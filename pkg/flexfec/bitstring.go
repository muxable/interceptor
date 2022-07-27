package flexfec

import (
	"encoding/binary"
	"fmt"

	"github.com/pion/rtp"
)

// Converts rtp packet to bitstring as per fec scheme-20
func ToBitString(pkt *rtp.Packet) (out []byte) {
	buf, err := pkt.Marshal()

	if err != nil {
		fmt.Println(err)
	}

	length := uint16(len(buf))

	// replace SN with length
	binary.BigEndian.PutUint16(buf[2:4], length)

	// remove SSRC
	bitstring := make([]byte, length-4)
	copy(bitstring[:8], buf[:8])
	copy(bitstring[8:], buf[12:])

	return bitstring
}

// Computes the xor of all the packets in the input array
func ToFecBitString(buf [][]byte) []byte {
	var buf_xor []byte
	buf_xor = append(buf[0])

	m := len(buf_xor)
	n := len(buf)

	for i := 1; i < n; i++ {
		for j := 0; j < m; j++ {
			// xor operation
			buf_xor[j] ^= buf[i][j]
		}
	}
	return buf_xor
}

// ------------------------------------------------------------

func GetBlockBitstring(packets *[]rtp.Packet) [][]byte {
	var bitStrings [][]byte

	for _, pkt := range *packets {
		bitStrings = append(bitStrings, ToBitString(&pkt))
	}

	return bitStrings
}

func PadBitStrings(bitstrings *[][]byte, length int) {
	maxSize := -1
	n := len(*bitstrings)

	for _, bitstring := range *bitstrings {
		currSize := len(bitstring)
		if maxSize < currSize {
			maxSize = currSize
		}
	}

	if maxSize < length {
		maxSize = length
		// fmt.Println("-------------------------max length updated------------------")
	}

	for i := 0; i < n; i++ {
		size := len((*bitstrings)[i])

		if size < maxSize {
			paddingSize := maxSize - size

			padding := make([]byte, paddingSize)
			(*bitstrings)[i] = append((*bitstrings)[i], padding...)
		}

	}

}
