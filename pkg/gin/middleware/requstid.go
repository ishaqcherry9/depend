package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ishaqcherry9/depend/pkg/krand"
)

var (
	ContextRequestIDKey = "request_id"
	HeaderXRequestIDKey = "X-Request-Id"
)

type RequestIDOption func(*requestIDOptions)

type requestIDOptions struct {
	contextRequestIDKey string
	headerXRequestIDKey string
}

func defaultRequestIDOptions() *requestIDOptions {
	return &requestIDOptions{
		contextRequestIDKey: ContextRequestIDKey,
		headerXRequestIDKey: HeaderXRequestIDKey,
	}
}

func (o *requestIDOptions) apply(opts ...RequestIDOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func (o *requestIDOptions) setRequestIDKey() {
	if o.contextRequestIDKey != ContextRequestIDKey {
		ContextRequestIDKey = o.contextRequestIDKey
	}
	if o.headerXRequestIDKey != HeaderXRequestIDKey {
		HeaderXRequestIDKey = o.headerXRequestIDKey
	}
}

func WithContextRequestIDKey(key string) RequestIDOption {
	return func(o *requestIDOptions) {
		if len(key) < 4 {
			return
		}
		o.contextRequestIDKey = key
	}
}

func WithHeaderRequestIDKey(key string) RequestIDOption {
	return func(o *requestIDOptions) {
		if len(key) < 4 {
			return
		}
		o.headerXRequestIDKey = key
	}
}

type CtxKeyString string

var RequestIDKey = CtxKeyString(ContextRequestIDKey)

func RequestID(opts ...RequestIDOption) gin.HandlerFunc {

	o := defaultRequestIDOptions()
	o.apply(opts...)
	o.setRequestIDKey()

	return func(c *gin.Context) {

		requestID := c.Request.Header.Get(HeaderXRequestIDKey)

		if requestID == "" {
			requestID = krand.String(krand.R_All, 10)
			c.Request.Header.Set(HeaderXRequestIDKey, requestID)
		}

		c.Set(ContextRequestIDKey, requestID)

		c.Writer.Header().Set(HeaderXRequestIDKey, requestID)

		c.Next()
	}
}

func GCtxRequestID(c *gin.Context) string {
	if v, isExist := c.Get(ContextRequestIDKey); isExist {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}

func GCtxRequestIDField(c *gin.Context) zap.Field {
	return zap.String(ContextRequestIDKey, GCtxRequestID(c))
}

func HeaderRequestID(c *gin.Context) string {
	return c.Request.Header.Get(HeaderXRequestIDKey)
}

func HeaderRequestIDField(c *gin.Context) zap.Field {
	return zap.String(HeaderXRequestIDKey, HeaderRequestID(c))
}

var RequestHeaderKey = "request_header_key"

func WrapCtx(c *gin.Context) context.Context {
	ctx := context.WithValue(c.Request.Context(), ContextRequestIDKey, c.GetString(ContextRequestIDKey))
	return context.WithValue(ctx, RequestHeaderKey, c.Request.Header)
}

func AdaptCtx(ctx context.Context) (*gin.Context, context.Context) {
	c, ok := ctx.(*gin.Context)
	if ok {
		ctx = WrapCtx(c)
	}
	return c, ctx
}

func GetFromCtx(ctx context.Context, key string) interface{} {
	return ctx.Value(key)
}

func CtxRequestID(ctx context.Context) string {
	v := ctx.Value(ContextRequestIDKey)
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

func CtxRequestIDField(ctx context.Context) zap.Field {
	return zap.String(ContextRequestIDKey, CtxRequestID(ctx))
}

func GetFromHeader(ctx context.Context, key string) string {
	header, ok := ctx.Value(RequestHeaderKey).(http.Header)
	if !ok {
		return ""
	}
	return header.Get(key)
}

func GetFromHeaders(ctx context.Context, key string) []string {
	header, ok := ctx.Value(RequestHeaderKey).(http.Header)
	if !ok {
		return []string{}
	}
	return header.Values(key)
}
