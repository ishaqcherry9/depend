package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ishaqcherry9/depend/pkg/errcode"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func newResp(code int, msg string, data interface{}) *Result {
	resp := &Result{
		Code: code,
		Msg:  msg,
	}

	if data == nil {
		resp.Data = &struct{}{}
	} else {
		resp.Data = data
	}

	return resp
}

var jsonContentType = []string{"application/json; charset=utf-8"}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

func writeJSON(c *gin.Context, code int, res interface{}) {
	c.Writer.WriteHeader(code)
	writeContentType(c.Writer, jsonContentType)
	err := json.NewEncoder(c.Writer).Encode(res)
	if err != nil {
		fmt.Printf("json encode error, err = %s\n", err.Error())
	}
}

func respJSONWithStatusCode(c *gin.Context, code int, msg string, data ...interface{}) {
	var firstData interface{}
	if len(data) > 0 {
		firstData = data[0]
	}
	resp := newResp(code, msg, firstData)

	writeJSON(c, code, resp)
}

func Output(c *gin.Context, code int, data ...interface{}) {
	switch code {
	case http.StatusOK:
		respJSONWithStatusCode(c, http.StatusOK, "ok", data...)
	case http.StatusBadRequest:
		respJSONWithStatusCode(c, http.StatusBadRequest, errcode.InvalidParams.Msg(), data...)
	case http.StatusUnauthorized:
		respJSONWithStatusCode(c, http.StatusUnauthorized, errcode.Unauthorized.Msg(), data...)
	case http.StatusForbidden:
		respJSONWithStatusCode(c, http.StatusForbidden, errcode.Forbidden.Msg(), data...)
	case http.StatusNotFound:
		respJSONWithStatusCode(c, http.StatusNotFound, errcode.NotFound.Msg(), data...)
	case http.StatusRequestTimeout:
		respJSONWithStatusCode(c, http.StatusRequestTimeout, errcode.Timeout.Msg(), data...)
	case http.StatusInternalServerError:
		respJSONWithStatusCode(c, http.StatusInternalServerError, errcode.InternalServerError.Msg(), data...)
	case http.StatusTooManyRequests:
		respJSONWithStatusCode(c, http.StatusTooManyRequests, errcode.LimitExceed.Msg(), data...)
	case http.StatusServiceUnavailable:
		respJSONWithStatusCode(c, http.StatusServiceUnavailable, errcode.ServiceUnavailable.Msg(), data...)

	default:
		respJSONWithStatusCode(c, code, http.StatusText(code), data...)
	}
}

func Out(c *gin.Context, err *errcode.Error, data ...interface{}) {
	code := err.ToHTTPCode()
	switch code {
	case http.StatusOK:
		respJSONWithStatusCode(c, http.StatusOK, "ok", data...)
	case http.StatusInternalServerError:
		respJSONWithStatusCode(c, http.StatusInternalServerError, err.Msg(), data...)
	case http.StatusBadRequest:
		respJSONWithStatusCode(c, http.StatusBadRequest, err.Msg(), data...)
	case http.StatusUnauthorized:
		respJSONWithStatusCode(c, http.StatusUnauthorized, err.Msg(), data...)
	case http.StatusForbidden:
		respJSONWithStatusCode(c, http.StatusForbidden, err.Msg(), data...)
	case http.StatusNotFound:
		respJSONWithStatusCode(c, http.StatusNotFound, err.Msg(), data...)
	case http.StatusRequestTimeout:
		respJSONWithStatusCode(c, http.StatusRequestTimeout, err.Msg(), data...)
	case http.StatusConflict:
		respJSONWithStatusCode(c, http.StatusConflict, err.Msg(), data...)
	case http.StatusTooManyRequests:
		respJSONWithStatusCode(c, http.StatusTooManyRequests, err.Msg(), data...)
	case http.StatusServiceUnavailable:
		respJSONWithStatusCode(c, http.StatusServiceUnavailable, err.Msg(), data...)

	default:
		respJSONWithStatusCode(c, http.StatusNotExtended, err.Msg(), data...)
	}
}

func respJSONWith200(c *gin.Context, code int, msg string, data ...interface{}) {
	var firstData interface{}
	if len(data) > 0 {
		firstData = data[0]
	}
	resp := newResp(code, msg, firstData)

	writeJSON(c, http.StatusOK, resp)
}

func Success(c *gin.Context, data ...interface{}) {
	respJSONWith200(c, 0, "ok", data...)
}

func Error(c *gin.Context, err *errcode.Error, data ...interface{}) {
	respJSONWith200(c, err.Code(), err.Msg(), data...)
}
