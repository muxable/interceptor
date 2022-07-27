package flexfec

import (
	"encoding/binary"
)

func ToFecHeaderLD(buf []byte, SN_base uint16, L, D uint8) (FecHeaderLD, []byte) {
	var fecheader FecHeaderLD
	// first 2 bits are neglected in FEC bit string and replaced by R and F
	fecheader.R = false
	fecheader.F = true
	fecheader.P = (buf[0] >> 5 & 0x1) > 0
	fecheader.X = (buf[0] >> 4 & 0x1) > 0
	fecheader.CC = uint8((buf[0] & uint8(0xF)))

	fecheader.M = (buf[1] >> 7 & 0x1) > 0
	fecheader.PTRecovery = buf[1] & 0x7F

	fecheader.LengthRecovery = binary.BigEndian.Uint16(buf[2:4])

	fecheader.TimestampRecovery = binary.BigEndian.Uint32(buf[4:8])

	// Check: SN_base, L, D
	fecheader.SN_base = SN_base
	fecheader.L = L
	fecheader.D = D
	return fecheader, buf[8:]
}

func ToFecHeaderFlexibleMask(buf []byte, SN_base, mask uint16, optionalMask1 uint32, optionalMask2 uint64) (FecHeaderFlexibleMask, []byte) {
	var fecheader FecHeaderFlexibleMask
	fecheader.R = false
	fecheader.F = false
	fecheader.P = (buf[0] >> 5 & 0x1) > 0
	fecheader.X = (buf[0] >> 4 & 0x1) > 0
	fecheader.CC = uint8((buf[0] & uint8(0xF)))

	fecheader.M = (buf[1] >> 7 & 0x1) > 0
	fecheader.PTRecovery = buf[1] & 0x7F

	fecheader.LengthRecovery = binary.BigEndian.Uint16(buf[2:4])

	fecheader.TimestampRecovery = binary.BigEndian.Uint32(buf[4:8])

	fecheader.SN_base = SN_base

	fecheader.Mask = mask

	if optionalMask1 > 0 {
		fecheader.OptionalMask1 = optionalMask1
		fecheader.K1 = true
	}

	if optionalMask2 > 0 {
		fecheader.OptionalMask2 = optionalMask2
		fecheader.K2 = true
	}
	return fecheader, buf[8:]
}
