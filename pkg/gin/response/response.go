package response

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ishaqcherry9/depend/pkg/errcode"
	jsoniter "github.com/ishaqcherry9/depend/pkg/json-iterator"
)

type Result struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	RequestId string      `json:"requestId"`
	Time      int64       `json:"time,omitempty"`
}

func GetRequestId(c *gin.Context) string {
	if v, isExist := c.Get("request_id"); isExist {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}

func newResp(code int, msg, requestId string, data interface{}) *Result {
	resp := &Result{
		Code:      code,
		Msg:       msg,
		RequestId: requestId,
	}

	if data == nil {
		resp.Data = &struct{}{}
	} else {
		resp.Data = data
	}

	return resp
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

func writeJSON(c *gin.Context, code int, res interface{}) {
	c.Writer.WriteHeader(code)

	jsonContentType := []string{"application/json; charset=utf-8"}
	writeContentType(c.Writer, jsonContentType)

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.NewEncoder(c.Writer).Encode(res)
	if err != nil {
		fmt.Printf("json encode error, err = %s\n", err.Error())
		http.Error(c.Writer, `{"code":500,"msg":"internal error"}`, http.StatusInternalServerError)
	}
}

func respJSONWithStatusCode(c *gin.Context, code int, msg string, data ...interface{}) {
	var firstData interface{}
	if len(data) > 0 {
		firstData = data[0]
	}

	resp := newResp(code, msg, GetRequestId(c), firstData)
	if code == http.StatusOK {
		resp.Time = time.Now().Unix()
	}
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

	resp := newResp(code, msg, GetRequestId(c), firstData)
	if code == 0 {
		resp.Time = time.Now().Unix()
	}
	writeJSON(c, http.StatusOK, resp)
}

func Success(c *gin.Context, data ...interface{}) {
	respJSONWith200(c, 0, "ok", data...)
}

// 返回码200，如果是4xx、5xx等请使用ErrorStatus
func Error(c *gin.Context, err *errcode.Error, data ...interface{}) {
	respJSONWith200(c, err.Code(), err.Msg(), data...)
}

// 返回码200，如果是4xx、5xx等请使用ErrorStatus
func ErrorWithCodeMsg(c *gin.Context, code int, msg string, data ...interface{}) {
	respJSONWith200(c, code, msg, data...)
}

// 如参数缺失http状态码应返400，授权失败返401，内部错误返500等。按HTTP协议语义返回，便于后续监控&可观测性建设。
func ErrorStatus(c *gin.Context, httpStatusCode int, err *errcode.Error, data ...interface{}) {
	var firstData interface{}
	if len(data) > 0 {
		firstData = data[0]
	}

	resp := newResp(err.Code(), err.Msg(), GetRequestId(c), firstData)
	writeJSON(c, httpStatusCode, resp)
}
