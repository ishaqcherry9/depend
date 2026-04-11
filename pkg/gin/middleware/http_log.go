package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/ishaqcherry9/depend/pkg/json-iterator"
	"github.com/ishaqcherry9/depend/pkg/logger"
)

const (
	maxLoggedBodyBytes = 10000
	truncatedSuffix    = "...[middleware content truncated]"
	omittedBodyHint    = "[binary or unsupported body omitted]"
)

// HttpLogMiddleware 记录接口请求/响应日志。
func HttpLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 校验
		if !validateHTTPLogContext(c) {
			return
		}

		// 2. 组装参数
		startAt := time.Now()
		meta, requestHeaders, requestBody, readErr := buildHTTPRequestLogParams(c)

		// 3. 记录请求
		logHTTPRequest(c, meta, requestHeaders, requestBody, readErr)

		// 4. defer 记录响应
		respWriter := &bodyLogWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = respWriter
		defer logHTTPResponse(c, meta, respWriter, startAt)

		// 5. 执行业务链路
		c.Next()
	}
}

// validateHTTPLogContext 校验日志中间件所需上下文。
func validateHTTPLogContext(c *gin.Context) bool {
	if c == nil {
		logger.Error(context.Background(), "context is nil", logger.String("log_tag", "middleware http start"))
		return false
	}
	if c.Request == nil {
		logger.Error(context.Background(), "request is nil", logger.String("log_tag", "middleware http start"))
		c.Next()
		return false
	}
	return true
}

// buildHTTPRequestLogParams 组装请求阶段日志参数。
func buildHTTPRequestLogParams(c *gin.Context) (httpLogMeta, http.Header, []byte, error) {
	defer func() {
		if r := recover(); r != nil {
			// 防御性兜底：日志参数组装失败时仅记录错误，避免中断主业务链路。
			logger.Error(c, "buildHTTPRequestLogParams panic", logger.Any("panic", r))
		}
	}()
	// 组装请求/响应共用日志字段。
	method := c.Request.Method
	uri := ""
	path := ""
	rawQuery := ""
	if c.Request.URL != nil {
		uri = c.Request.URL.RequestURI()
		path = c.Request.URL.Path
		rawQuery = c.Request.URL.RawQuery
	}
	// siteID, siteName := getSiteMeta()
	meta := httpLogMeta{
		method:   method,
		uri:      uri,
		path:     path,
		rawQuery: rawQuery,
		userID:   GetCustomerUidByCtx(c),
		clientIP: getClientIP(c),
		serverIP: getServerIP(c),
		// siteID:   siteID,
		// siteName: siteName,
	}
	requestHeaders := c.Request.Header
	// 读取请求体内容。
	var (
		requestBody []byte
		readErr     error
	)
	if shouldLogBody(requestHeaders.Get("Content-Type")) {
		requestBody, readErr = readAndRestoreBody(c)
	}
	return meta, requestHeaders, requestBody, readErr
}

// logHTTPRequest 打印请求阶段日志。
func logHTTPRequest(ctx context.Context, meta httpLogMeta, headers http.Header, requestBody []byte, readErr error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "logHTTPRequest panic", logger.Any("panic", r))
		}
	}()

	logger.Info(
		ctx,
		"http_request",
		logger.String("method", meta.method),
		logger.String("uri", meta.uri),
		logger.Any("path", meta.path),
		logger.Any("query", meta.rawQuery),
		logger.Any("headers", headers),
		logger.Any("user_id", meta.userID),
		logger.String("client_ip", meta.clientIP),
		logger.String("server_ip", meta.serverIP),
		// logger.String("site_id", meta.siteID),
		// logger.String("site_name", meta.siteName),
		logger.String("log_tag", "middleware http start"),
		logger.Any("body", formatLoggedBody(headers, requestBody)),
		logger.Any("body_length", len(requestBody)),
		logger.Any("body_read_error", formatError(readErr)),
	)
}

// logHTTPResponse 打印响应阶段日志。
func logHTTPResponse(c *gin.Context, meta httpLogMeta, respWriter *bodyLogWriter, startAt time.Time) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(c, "logHTTPResponse panic", logger.Any("panic", r))
		}
	}()
	if c == nil || respWriter == nil {
		logger.Error(c, "logHTTPResponse invalid context or response writer")
		return
	}

	elapsedMs := time.Since(startAt).Milliseconds()
	errMessage := ""
	if len(c.Errors) > 0 {
		errMessage = c.Errors.String()
	}

	statusCode := respWriter.Status()
	respHeaders := respWriter.Header()

	bodyCode := extractBodyCode(respHeaders, respWriter.body.Bytes())
	body := formatLoggedBody(respHeaders, respWriter.body.Bytes())
	contentBytes, contentErr := jsoniter.Marshal(map[string]interface{}{
		"status_code":     statusCode,
		"body_code":       bodyCode,
		"body_length":     respWriter.body.Len(),
		"elapsed_time_ms": elapsedMs,
	})
	content := ""
	if contentErr != nil {
		logger.Error(c, "logHTTPResponse content marshal error", logger.Any("error", contentErr))
	} else {
		content = string(contentBytes)
	}
	logger.Info(
		c,
		"http_response",
		logger.String("method", meta.method),
		logger.String("uri", meta.uri),
		logger.Any("path", meta.path),
		logger.Any("query", meta.rawQuery),
		logger.Any("headers", respHeaders),
		logger.Any("user_id", meta.userID),
		logger.String("client_ip", meta.clientIP),
		logger.String("server_ip", meta.serverIP),
		// logger.String("site_id", meta.siteID),
		// logger.String("site_name", meta.siteName),
		logger.Any("status_code", statusCode),
		logger.Any("elapsed_time_ms", elapsedMs),
		logger.Any("body", body),
		logger.Any("body_code", bodyCode),
		logger.Any("body_length", respWriter.body.Len()),
		logger.String("error", errMessage),
		logger.String("log_tag", "middleware http end"),
		logger.String("content", content),
		// logger.Any("content", map[string]interface{}{
		// 	"status_code":     statusCode,
		// 	"body_code":       bodyCode,
		// 	"body_length":     respWriter.body.Len(),
		// 	"elapsed_time_ms": elapsedMs,
		// }),
	)
}

// readAndRestoreBody 读取请求体并重置流，确保后续处理仍可读取 body。
func readAndRestoreBody(c *gin.Context) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(c, "readAndRestoreBody panic", logger.Any("panic", err))
		}
	}()
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error(c, "readAndRestoreBody read body error", logger.Any("error", err))
		return nil, err
	}

	// 读完后重置流，保证后续 handler/绑定逻辑仍可再次读取请求体。
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// formatLoggedBody 根据请求/响应头中的内容类型格式化日志 body，并在超长时截断。
func formatLoggedBody(headers http.Header, body []byte) string {
	if len(body) == 0 {
		return ""
	}
	contentType := ""
	if headers != nil {
		contentType = headers.Get("Content-Type")
	}
	if !shouldLogBody(contentType) {
		return omittedBodyHint
	}

	if len(body) <= maxLoggedBodyBytes {
		return string(body)
	}

	return string(body[:maxLoggedBodyBytes]) + truncatedSuffix
}

// shouldLogBody 判断当前 content-type 是否适合记录 body。
// 参考 PHP 中间件日志策略，仅记录可安全解析的文本内容类型。
func shouldLogBody(contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct == "" {
		return false
	}
	if idx := strings.Index(ct, ";"); idx > -1 {
		ct = strings.TrimSpace(ct[:idx])
	}

	switch ct {
	case "application/json", "application/x-www-form-urlencoded", "text/plain":
		return true
	default:
		return false
	}
}

// formatError 将错误对象格式化为日志字段内容。
func formatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// extractBodyCode 从 JSON 响应中提取 code 字段，便于统一统计。
func extractBodyCode(headers http.Header, body []byte) interface{} {
	if len(body) == 0 {
		return nil
	}
	contentType := ""
	if headers != nil {
		contentType = headers.Get("Content-Type")
	}
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}
	if code, ok := data["code"]; ok {
		return code
	}
	return nil
}

// getServerIP 获取当前处理请求的本机服务 IP。
func getServerIP(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}

	// 使用当前请求连接到的本地地址（类似 SERVER_ADDR）
	localAddr := c.Request.Context().Value(http.LocalAddrContextKey)
	if addr, ok := localAddr.(net.Addr); ok && addr != nil {
		if tcpAddr, ok := addr.(*net.TCPAddr); ok && tcpAddr.IP != nil {
			return tcpAddr.IP.String()
		}
		host, _, err := net.SplitHostPort(addr.String())
		if err == nil && host != "" {
			return host
		}
		return addr.String()
	}

	return ""
}

// getClientIP 获取客户端真实 IP，优先使用 gin 封装。
func getClientIP(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if ip := strings.TrimSpace(c.ClientIP()); ip != "" {
		return ip
	}
	if host, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(c.Request.RemoteAddr)
}

// // getSiteMeta 获取站点元信息（site_id、site_name）。
// func getSiteMeta() (string, string) {
// 	siteID := strings.TrimSpace(os.Getenv("site_id"))
// 	siteName := strings.TrimSpace(os.Getenv("SITE_NAME"))

// 	cfg := config.Get()
// 	if cfg != nil {
// 		if siteID == "" {
// 			siteID = strings.TrimSpace(cfg.App.SiteID)
// 		}
// 		if siteName == "" {
// 			siteName = strings.TrimSpace(cfg.App.Name)
// 		}
// 	}

// 	return siteID, siteName
// }

// // getUserID 提取用户ID，优先请求头。
// func getUserID(c *gin.Context) string {
// 	if c == nil {
// 		return ""
// 	}
// 	return GetCustomerUidByCtx(c)
// }

// httpLogMeta 统一承载请求/响应公共日志字段。
type httpLogMeta struct {
	method   string
	uri      string
	path     string
	rawQuery string
	userID   string
	clientIP string
	serverIP string
	// siteID   string
	// siteName string
}

// bodyLogWriter1 用于缓存响应体内容。
type bodyLogWriter1 struct {
	gin.ResponseWriter              // gin 原始响应写入器
	body               bytes.Buffer // 用于缓存响应体内容
}

// Write 写入响应并缓存响应体内容。
func (w *bodyLogWriter1) Write(data []byte) (int, error) {
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// WriteString 写入字符串响应并缓存响应体内容。
func (w *bodyLogWriter1) WriteString(s string) (int, error) {
	_, _ = w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
