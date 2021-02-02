package rpcode

import (
	"strconv"

	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc/status"
)

type rpcStatus struct {
	Code    string     `json:"code,omitempty"`
	Message string     `json:"message,omitempty"`
	Details []*any.Any `json:"details,omitempty"`
}

// ExtractCodes cause from error to ecode.
func ExtractCode(e error) *rpcStatus {
	if e == nil {
		return &rpcStatus{
			Code:    "0",
			Message: "OK",
		}
	}
	// todo 不想做code类型转换，所以全部用grpc标准码处理
	gst, _ := status.FromError(e)
	return &rpcStatus{
		Code:    strconv.Itoa(int(gst.Code())),
		Message: gst.Message(),
		Details: make([]*any.Any, 0),
	}
}
