package errcode

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const ToHTTPCodeLabel = "[standard http code]"

var errCodes = map[int]*Error{}
var httpErrCodes = map[int]string{}

type Error struct {
	code         int
	msg          string
	details      []string
	needHTTPCode bool
}

func NewError(code int, msg string, details ...string) *Error {
	if v, ok := errCodes[code]; ok {
		panic(fmt.Sprintf(`http error code = %d already exists, please define a new error code,
msg1 = %s
msg2 = %s
`, code, v.Msg(), msg))
	}

	httpErrCodes[code] = msg
	e := &Error{code: code, msg: msg, details: details}
	errCodes[code] = e
	return e
}

func (e *Error) Err(msg ...string) error {
	message := e.msg
	if len(msg) > 0 {
		message = strings.Join(msg, ", ")
	}

	if len(e.details) == 0 {
		return fmt.Errorf("code = %d, msg = %s", e.code, message)
	}
	return fmt.Errorf("code = %d, msg = %s, details = %v", e.code, message, e.details)
}

func (e *Error) ErrToHTTP(msg ...string) error {
	message := e.msg
	if len(msg) > 0 {
		message = strings.Join(msg, ", ")
	}

	if len(e.details) == 0 {
		return fmt.Errorf("code = %d, msg = %s%s", e.code, message, ToHTTPCodeLabel)
	}
	return fmt.Errorf("code = %d, msg = %s, details = %v%s", e.code, message, strings.Join(e.details, ", "), ToHTTPCodeLabel)
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Msg() string {
	return e.msg
}

func (e *Error) NeedHTTPCode() bool {
	return e.needHTTPCode
}

func (e *Error) Details() []string {
	return e.details
}

func (e *Error) WithDetails(details ...string) *Error {
	newError := &Error{code: e.code, msg: e.msg}
	newError.msg += ", " + strings.Join(details, ", ")
	return newError
}

func (e *Error) RewriteMsg(msg string) *Error {
	return &Error{code: e.code, msg: msg}
}

func (e *Error) WithOutMsg(msg string) *Error {
	return &Error{code: e.code, msg: msg}
}

func (e *Error) WithOutMsgI18n(langMsg map[int]map[string]string, lang string) *Error {
	if i18nMsg, ok := langMsg[e.Code()]; ok {
		if msg, ok2 := i18nMsg[lang]; ok2 {
			return &Error{code: e.code, msg: msg}
		}
	}

	return &Error{code: e.code, msg: e.msg}
}

func (e *Error) ToHTTPCode() int {
	switch e.Code() {
	case Success.Code():
		return http.StatusOK
	case InternalServerError.Code():
		return http.StatusInternalServerError
	case InvalidParams.Code():
		return http.StatusBadRequest
	}

	switch e.Code() {
	case Unauthorized.Code(), PermissionDenied.Code():
		return http.StatusUnauthorized
	case TooManyRequests.Code(), LimitExceed.Code():
		return http.StatusTooManyRequests
	case Forbidden.Code(), AccessDenied.Code():
		return http.StatusForbidden
	case NotFound.Code():
		return http.StatusNotFound
	case TooEarly.Code():
		return http.StatusTooEarly
	case Timeout.Code(), DeadlineExceeded.Code():
		return http.StatusRequestTimeout
	case MethodNotAllowed.Code():
		return http.StatusMethodNotAllowed
	case ServiceUnavailable.Code():
		return http.StatusServiceUnavailable
	case Unimplemented.Code():
		return http.StatusNotImplemented
	case StatusBadGateway.Code():
		return http.StatusBadGateway
	}

	return http.StatusInternalServerError
}

func ParseError(err error) *Error {
	if err == nil {
		return Success
	}

	outError := &Error{
		code: -1,
		msg:  "unknown error",
	}

	splits := strings.Split(err.Error(), ", msg = ")
	codeStr := strings.ReplaceAll(splits[0], "code = ", "")
	code, er := strconv.Atoi(codeStr)
	if er != nil {
		return outError
	}

	if e, ok := errCodes[code]; ok {
		if len(splits) > 1 {
			outError.code = code
			outError.msg = splits[1]
			outError.needHTTPCode = strings.Contains(err.Error(), ToHTTPCodeLabel)
			return outError
		}
		return e
	}

	return outError
}

func GetErrorCode(err error) int {
	e := ParseError(err)
	if e.needHTTPCode {
		return e.ToHTTPCode()
	}
	return e.Code()
}
