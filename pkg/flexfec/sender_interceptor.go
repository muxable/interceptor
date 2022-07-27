package flexfec

import (
	"fmt"

	"github.com/pion/interceptor"
)

type SenderInterceptor struct {
	interceptor.NoOp
}

func NewSenderInterceptor() interceptor.Interceptor {
	return &SenderInterceptor{}
}

func (i *SenderInterceptor) BindLocalStream(info *interceptor.StreamInfo, writer interceptor.RTPWriter) interceptor.RTPWriter {
	fmt.Println("BindLocalStream called")
	return writer
}
