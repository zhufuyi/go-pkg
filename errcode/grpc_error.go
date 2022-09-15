package errcode

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCStatus grpc 状态
type GRPCStatus struct {
	status *status.Status
}

var statusCodes = map[codes.Code]string{}

// NewGRPCStatus 新建一个status
func NewGRPCStatus(code codes.Code, msg string) *GRPCStatus {
	if v, ok := statusCodes[code]; ok {
		panic(fmt.Sprintf("grpc status code = %d already exists, please replace with a new error code, old msg = %s", code, v))
	} else {
		statusCodes[code] = msg
	}

	return &GRPCStatus{
		status: status.New(code, msg),
	}
}

// Detail error details
type Detail struct {
	key string
	val interface{}
}

// String detail key-value
func (d Detail) String() string {
	return fmt.Sprintf("%s: {%v}", d.key, d.val)
}

// Any type
func Any(key string, val interface{}) Detail {
	return Detail{
		key: key,
		val: val,
	}
}

// RPCErr rpc error
func RPCErr(g *GRPCStatus, details ...Detail) error {
	var dts []string
	for _, detail := range details {
		dts = append(dts, detail.String())
	}

	return status.Errorf(g.status.Code(), "%s details = %v", g.status.Message(), dts)
}

// Err return error
func (g *GRPCStatus) Err(details ...Detail) error {
	var dts []string
	for _, detail := range details {
		dts = append(dts, detail.String())
	}
	return status.Errorf(g.status.Code(), "%s details = %v", g.status.Message(), dts)
}

// ToRPCCode 转换为RPC识别的错误码，避免返回Unknown状态码
func ToRPCCode(code int) codes.Code {
	var statusCode codes.Code

	switch code {
	case InternalServerError.code:
		statusCode = codes.Internal
	case InvalidParams.code:
		statusCode = codes.InvalidArgument
	case Unauthorized.code:
		statusCode = codes.Unauthenticated
	case NotFound.code:
		statusCode = codes.NotFound
	case DeadlineExceeded.code:
		statusCode = codes.DeadlineExceeded
	case AccessDenied.code:
		statusCode = codes.PermissionDenied
	case LimitExceed.code:
		statusCode = codes.ResourceExhausted
	case MethodNotAllowed.code:
		statusCode = codes.Unimplemented
	default:
		statusCode = codes.Unknown
	}

	return statusCode
}
