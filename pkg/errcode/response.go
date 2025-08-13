package errcode

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var SkipResponse = errors.New("skip response")

type Responser interface {
	Success(ctx *gin.Context, data interface{})
	ParamError(ctx *gin.Context, err error)
	Error(ctx *gin.Context, err error) bool
}

func NewResponser(httpErrors []*Error) Responser {
	httpErrorsMap := make(map[int]*Error)

	for _, httpError := range httpErrors {
		if httpError == nil {
			continue
		}
		httpErrorsMap[httpError.Code()] = httpError
	}

	return &defaultResponse{
		httpErrors: httpErrorsMap,
	}
}

type defaultResponse struct {
	httpErrors map[int]*Error
}

func (resp *defaultResponse) response(c *gin.Context, respStatus, code int, msg string, data interface{}) {
	c.JSON(respStatus, map[string]interface{}{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

func (resp *defaultResponse) Success(c *gin.Context, data interface{}) {
	resp.response(c, http.StatusOK, 0, "ok", data)
}

func (resp *defaultResponse) ParamError(c *gin.Context, _ error) {
	resp.response(c, http.StatusOK, InvalidParams.Code(), InvalidParams.Msg(), struct{}{})
}

func (resp *defaultResponse) Error(c *gin.Context, err error) bool {

	return resp.handleHTTPError(c, err)
}

func (resp *defaultResponse) handleHTTPError(c *gin.Context, err error) bool {
	e := ParseError(err)

	switch e.Code() {
	case InternalServerError.Code(), http.StatusInternalServerError:
		resp.response(c, http.StatusInternalServerError, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), struct{}{})
		return true
	case ServiceUnavailable.Code(), http.StatusServiceUnavailable:
		resp.response(c, http.StatusServiceUnavailable, http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable), struct{}{})
		return true
	}

	if e.needHTTPCode {
		msg := strings.ReplaceAll(e.msg, ToHTTPCodeLabel, "")
		resp.response(c, e.ToHTTPCode(), e.code, msg, struct{}{})
		return true
	}

	if resp.isUserDefinedHTTPErrorCode(c, e.Code()) {
		return true
	}

	resp.response(c, http.StatusOK, e.code, e.msg, struct{}{})
	return false
}

func (resp *defaultResponse) isUserDefinedHTTPErrorCode(c *gin.Context, errCode int) bool {
	if v, ok := resp.httpErrors[errCode]; ok {
		httpCode := v.ToHTTPCode()
		msg := http.StatusText(httpCode)
		if msg == "" {
			msg = "unknown error"
		}
		resp.response(c, httpCode, httpCode, msg, struct{}{})
		return true
	}
	return false
}
