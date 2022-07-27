package flexfec

import (
	"fmt"

	"github.com/pion/interceptor"
)

type ReceiverInterceptor struct {
	interceptor.NoOp
}

func NewReceiverInterceptor() interceptor.Interceptor {
	return &ReceiverInterceptor{}
}

func (i *ReceiverInterceptor) BindRemoteStream(info *interceptor.StreamInfo, reader interceptor.RTPReader) interceptor.RTPReader {
	fmt.Println("BindRemoteStream called")
	return reader
}
