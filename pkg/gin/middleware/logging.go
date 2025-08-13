package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	defaultMaxLength = 300

	defaultLogger, _ = zap.NewProduction()

	defaultIgnoreRoutes = map[string]struct{}{
		"/ping":   {},
		"/pong":   {},
		"/health": {},
	}

	printErrorBySpecifiedCodes = map[int]bool{
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
	}

	emptyBody   = []byte("")
	contentMark = []byte(" ...... ")
)

type Option func(*options)

func defaultOptions() *options {
	return &options{
		maxLength:     defaultMaxLength,
		log:           defaultLogger,
		ignoreRoutes:  defaultIgnoreRoutes,
		requestIDFrom: 0,
	}
}

type options struct {
	maxLength     int
	log           *zap.Logger
	ignoreRoutes  map[string]struct{}
	requestIDFrom int
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithMaxLen(maxLen int) Option {
	return func(o *options) {
		if maxLen < len(contentMark) {
			panic("maxLen should be greater than or equal to 8")
		}
		o.maxLength = maxLen
	}
}

func WithLog(log *zap.Logger) Option {
	return func(o *options) {
		if log != nil {
			o.log = log
		}
	}
}

func WithIgnoreRoutes(routes ...string) Option {
	return func(o *options) {
		for _, route := range routes {
			o.ignoreRoutes[route] = struct{}{}
		}
	}
}

func WithPrintErrorByCodes(code ...int) Option {
	return func(o *options) {
		for _, c := range code {
			printErrorBySpecifiedCodes[c] = true
		}
	}
}

func WithRequestIDFromContext() Option {
	return func(o *options) {
		o.requestIDFrom = 1
	}
}

func WithRequestIDFromHeader() Option {
	return func(o *options) {
		o.requestIDFrom = 2
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func getResponseBody(buf *bytes.Buffer, maxLen int) []byte {
	l := buf.Len()
	if l == 0 {
		return []byte("")
	} else if l > maxLen {
		l = maxLen
	}

	body := make([]byte, l)
	n, _ := buf.Read(body)
	if n == 0 {
		return emptyBody
	} else if n < maxLen {
		return body[:n-1]
	}
	return append(body[:maxLen-len(contentMark)], contentMark...)
}

func getRequestBody(buf *bytes.Buffer, maxLen int) []byte {
	l := buf.Len()
	if l == 0 {
		return []byte("")
	} else if l < maxLen {
		return buf.Bytes()
	}

	body := make([]byte, maxLen)
	copy(body, buf.Bytes())
	return append(body[:maxLen-len(contentMark)], contentMark...)
}

func Logging(opts ...Option) gin.HandlerFunc {
	o := defaultOptions()
	o.apply(opts...)

	return func(c *gin.Context) {
		start := time.Now()

		if _, ok := o.ignoreRoutes[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		buf := bytes.Buffer{}
		_, _ = buf.ReadFrom(c.Request.Body)
		sizeField := zap.Skip()
		bodyField := zap.Skip()
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch || c.Request.Method == http.MethodDelete {
			sizeField = zap.Int("size", buf.Len())
			bodyField = zap.ByteString("body", getRequestBody(&buf, o.maxLength))
		}

		reqID := ""
		if o.requestIDFrom == 1 {
			if v, isExist := c.Get(ContextRequestIDKey); isExist {
				if requestID, ok := v.(string); ok {
					reqID = requestID
				}
			}
		} else if o.requestIDFrom == 2 {
			reqID = c.Request.Header.Get(HeaderXRequestIDKey)
		}
		reqIDField := zap.Skip()
		if reqID != "" {
			reqIDField = zap.String(ContextRequestIDKey, reqID)
		}

		o.log.Info("<<<<",
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			sizeField,
			bodyField,
			reqIDField,
		)

		c.Request.Body = io.NopCloser(&buf)

		newWriter := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = newWriter

		c.Next()

		httpCode := c.Writer.Status()
		fields := []zap.Field{
			zap.Int("code", httpCode),
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.Path),
			zap.Int64("time_us", time.Since(start).Microseconds()),
			zap.Int("size", newWriter.body.Len()),
			zap.ByteString("body", getResponseBody(newWriter.body, o.maxLength)),
			reqIDField,
		}
		if printErrorBySpecifiedCodes[httpCode] {
			o.log.WithOptions(zap.AddStacktrace(zap.PanicLevel)).Error(">>>>", fields...)
		} else {
			o.log.Info(">>>>", fields...)
		}
	}
}

func SimpleLog(opts ...Option) gin.HandlerFunc {
	o := defaultOptions()
	o.apply(opts...)

	return func(c *gin.Context) {
		start := time.Now()

		if _, ok := o.ignoreRoutes[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		reqID := ""
		if o.requestIDFrom == 1 {
			if v, isExist := c.Get(ContextRequestIDKey); isExist {
				if requestID, ok := v.(string); ok {
					reqID = requestID
				}
			}
		} else if o.requestIDFrom == 2 {
			reqID = c.Request.Header.Get(HeaderXRequestIDKey)
		}
		reqIDField := zap.Skip()
		if reqID != "" {
			reqIDField = zap.String(ContextRequestIDKey, reqID)
		}

		c.Next()

		httpCode := c.Writer.Status()
		fields := []zap.Field{
			zap.Int("code", httpCode),
			zap.String("method", c.Request.Method),
			zap.String("url", c.Request.URL.String()),
			zap.Int64("time_us", time.Since(start).Microseconds()),
			zap.Int("size", c.Writer.Size()),
			reqIDField,
		}
		if printErrorBySpecifiedCodes[httpCode] {
			o.log.WithOptions(zap.AddStacktrace(zap.PanicLevel)).Error("Gin response", fields...)
		} else {
			o.log.Info("Gin response", fields...)
		}
	}
}
