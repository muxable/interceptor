package flexfec

import (
	"fmt"

	"github.com/pion/interceptor"
	"github.com/pion/rtp"
)

type ReceiverInterceptor struct {
	interceptor.NoOp
	BUFFER map[Key]rtp.Packet
	L      uint8
	D      uint8
	r      bool
	f      bool
}

func NewReceiverInterceptor() interceptor.Interceptor {
	return &ReceiverInterceptor{}
}

func (i *ReceiverInterceptor) BindRemoteStream(info *interceptor.StreamInfo, reader interceptor.RTPReader) interceptor.RTPReader {
	fmt.Println("BindRemoteStream called")
	return reader
}
