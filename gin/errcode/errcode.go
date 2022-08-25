package errcode

import (
	"fmt"
	"net/http"
)

// Error 错误
type Error struct {
	// 错误码
	code int
	// 错误消息
	msg string
	// 详细信息
	details []string
}

var codes = map[int]string{}

// NewError 创建新错误信息
func NewError(code int, msg string) *Error {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("错误码 %d 已经存在，请更换一个", code))
	}
	codes[code] = msg
	return &Error{code: code, msg: msg}
}

// String 打印错误
func (e *Error) Error() string {
	return fmt.Sprintf("错误码：%d, 错误信息:：%s", e.Code(), e.Msg())
}

// Code 错误码
func (e *Error) Code() int {
	return e.code
}

// Msg 错误信息
func (e *Error) Msg() string {
	return e.msg
}

// Msgf 附加信息
func (e *Error) Msgf(args []interface{}) string {
	return fmt.Sprintf(e.msg, args...)
}

// Details 错误详情
func (e *Error) Details() []string {
	return e.details
}

// WithDetails 携带附加错误详情
func (e *Error) WithDetails(details ...string) *Error {
	newError := *e
	newError.details = []string{}
	newError.details = append(newError.details, details...)

	return &newError
}

// StatusCode 对应http错误码
func (e *Error) StatusCode() int {
	switch e.Code() {
	case Success.Code():
		return http.StatusOK
	case InternalServerError.Code():
		return http.StatusInternalServerError
	case InvalidParams.Code():
		return http.StatusBadRequest
	case Unauthorized.Code():
		return http.StatusUnauthorized
	case TooManyRequests.Code(), LimitExceed.Code():
		return http.StatusTooManyRequests
	case Forbidden.Code():
		return http.StatusForbidden
	case NotFound.Code():
		return http.StatusNotFound
	case Timeout.Code():
		return http.StatusRequestTimeout
	}

	return http.StatusInternalServerError
}
