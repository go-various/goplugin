package goplugin

import (
	"google.golang.org/grpc"
	"math"
)

var largeMsgGRPCCallOpts []grpc.CallOption = []grpc.CallOption{
	grpc.MaxCallSendMsgSize(math.MaxInt32),
	grpc.MaxCallRecvMsgSize(math.MaxInt32),
}
