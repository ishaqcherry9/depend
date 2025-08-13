package nacos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ishaqcherry9/depend/pkg/cgreq"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	SmallBufferSize      = 4096
	MediumBufferSize     = 16384
	LargeBufferSize      = 65536
	ExtraLargeBufferSize = 131072
)

var (
	requestPool = &sync.Pool{
		New: func() interface{} {
			return &fasthttp.Request{}
		},
	}

	responsePool = &sync.Pool{
		New: func() interface{} {
			return &fasthttp.Response{}
		},
	}

	smallBufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, SmallBufferSize)
		},
	}

	mediumBufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, MediumBufferSize)
		},
	}

	largeBufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, LargeBufferSize)
		},
	}

	extraLargeBufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, ExtraLargeBufferSize)
		},
	}
)

type TieredBufferManager struct{}

func (tbm *TieredBufferManager) GetBuffer(size int) ([]byte, func()) {
	var pool *sync.Pool

	switch {
	case size <= SmallBufferSize:
		pool = smallBufferPool
	case size <= MediumBufferSize:
		pool = mediumBufferPool
	case size <= LargeBufferSize:
		pool = largeBufferPool
	case size <= ExtraLargeBufferSize:
		pool = extraLargeBufferPool
	default:
		return make([]byte, size), func() {}
	}

	buffer := pool.Get().([]byte)
	if cap(buffer) >= size {
		return buffer[:size], func() { pool.Put(buffer[:0]) }
	}

	directBuffer := make([]byte, size)
	return directBuffer, func() { pool.Put(buffer[:0]) }
}

var tieredBufferManager = &TieredBufferManager{}

const (
	maxRetries      = 3
	defaultTimeout  = 30 * time.Second
	backoffDuration = 1 * time.Second
)

func (sc *ServiceClient) Get(feign *cgreq.CgReq) ([]byte, error) {
	return sc.doWithRetry(feign.Context, "GET", feign)
}

func (sc *ServiceClient) Post(feign *cgreq.CgReq) ([]byte, error) {
	return sc.doWithRetry(feign.Context, "POST", feign)
}

func (sc *ServiceClient) Put(feign *cgreq.CgReq) ([]byte, error) {
	return sc.doWithRetry(feign.Context, "PUT", feign)
}

func (sc *ServiceClient) Delete(feign *cgreq.CgReq) ([]byte, error) {
	return sc.doWithRetry(feign.Context, "DELETE", feign)
}

func (sc *ServiceClient) doWithRetry(ctx context.Context, method string, feign *cgreq.CgReq) ([]byte, error) {

	failedAddresses := make(map[string]bool)
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {

		currentAddresses := sc.GetAddresses()
		if len(currentAddresses) == 0 {
			return nil, fmt.Errorf("no available addresses for service %s", sc.GetServiceName())
		}

		availableAddresses := sc.filterAvailableAddresses(currentAddresses, failedAddresses)
		if len(availableAddresses) == 0 {
			return nil, fmt.Errorf("all addresses failed for service %s: %w", sc.GetServiceName(), lastErr)
		}

		selectedAddress := sc.selectAddressWithLoadBalance(availableAddresses)

		if attempt > 0 {
			logger.Warnf("[ServiceClient] %s retry attempt %d, selected address: %s",
				sc.GetServiceName(), attempt, selectedAddress)
		} else {
			logger.Debugf("[ServiceClient] %s first attempt, selected address: %s",
				sc.GetServiceName(), selectedAddress)
		}

		result, err := sc.doFastHTTPRequest(ctx, method, selectedAddress, feign)
		if err == nil {
			if attempt > 0 {
				logger.Infof("[ServiceClient] %s request succeeded after %d retries",
					sc.GetServiceName(), attempt)
			}
			return result, nil
		}

		lastErr = err
		logger.Warnf("[ServiceClient] %s request failed to %s (attempt %d): %v",
			sc.GetServiceName(), selectedAddress, attempt+1, err)

		if !isRetryableError(err) {
			logger.Debugf("[ServiceClient] %s error is not retryable: %v",
				sc.GetServiceName(), err)
			return nil, err
		}

		failedAddresses[selectedAddress] = true

		if attempt < maxRetries {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil, fmt.Errorf("all retries failed for service %s: %w", sc.GetServiceName(), lastErr)
}

func (sc *ServiceClient) filterAvailableAddresses(currentAddresses []string, failedAddresses map[string]bool) []string {
	availableAddresses := make([]string, 0, len(currentAddresses))

	for _, addr := range currentAddresses {
		if !failedAddresses[addr] {
			availableAddresses = append(availableAddresses, addr)
		}
	}

	return availableAddresses
}

func (sc *ServiceClient) selectAddressWithLoadBalance(addresses []string) string {
	if len(addresses) == 0 {
		return ""
	}
	index := sc.counter.Add(1) % uint64(len(addresses))
	return addresses[index]
}

func (sc *ServiceClient) doFastHTTPRequest(ctx context.Context, method, address string, feign *cgreq.CgReq) ([]byte, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	req := requestPool.Get().(*fasthttp.Request)
	resp := responsePool.Get().(*fasthttp.Response)

	defer func() {
		req.Reset()
		resp.Reset()
		requestPool.Put(req)
		responsePool.Put(resp)
	}()

	fullURL := fmt.Sprintf("http://%s%s", address, feign.Path)
	req.SetRequestURI(fullURL)
	req.Header.SetMethod(method)

	timeout := feign.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, ctx.Err()
		}
		if remaining < timeout {
			return nil, ctx.Err()
		}
	}

	if feign.Body != nil && (method == "POST" || method == "PUT") {
		bodyBytes, err := json.Marshal(feign.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		req.SetBody(bodyBytes)
		req.Header.SetContentType("application/json")
	}

	if err := sc.setupHeaders(req, feign); err != nil {
		return nil, fmt.Errorf("failed to setup headers: %w", err)
	}

	err := fasthttp.DoTimeout(req, resp, timeout)
	if err != nil {
		return nil, fmt.Errorf("fasthttp request failed: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d: %s", statusCode, string(resp.Body()))
	}

	body := resp.Body()
	bodyLen := len(body)

	buffer, releaseFunc := tieredBufferManager.GetBuffer(bodyLen)
	defer releaseFunc()

	copy(buffer, body)

	result := make([]byte, bodyLen)
	copy(result, buffer)

	return result, nil
}

func (sc *ServiceClient) setupHeaders(req *fasthttp.Request, feign *cgreq.CgReq) error {
	req.Header.Set("User-Agent", "cg-http-client/1.0")
	req.Header.Set("Accept", "application/json")

	for k, v := range feign.Headers {
		req.Header.Set(k, v)
	}

	if err := sc.setTraceHeaders(req, feign.Context); err != nil {
		return err
	}

	return nil
}

func (sc *ServiceClient) setTraceHeaders(req *fasthttp.Request, ctx context.Context) error {

	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		req.Header.Set("X-Trace-ID", traceID)
	}

	if userInfo := getUserInfoFromContext(ctx); userInfo != "" {
		req.Header.Set("X-User-Info", userInfo)
	}

	if stepID := getStepIDFromContext(ctx); stepID != "" {
		req.Header.Set("X-Step-ID", stepID)
	}

	return nil
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, fasthttp.ErrTimeout), errors.Is(err, fasthttp.ErrNoFreeConns):
		return true
	}

	errStr := err.Error()

	return !strings.Contains(errStr, "HTTP error 4")
}

func getTraceIDFromContext(ctx context.Context) string {
	if value := ctx.Value("traceID"); value != nil {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}
	return ""
}

func getUserInfoFromContext(ctx context.Context) string {
	if value := ctx.Value("userInfo"); value != nil {
		if userInfo, ok := value.(string); ok {
			return userInfo
		}
	}
	return ""
}

func getStepIDFromContext(ctx context.Context) string {
	if value := ctx.Value("stepID"); value != nil {
		if stepID, ok := value.(string); ok {
			return stepID
		}
	}
	return ""
}

func warmupPools() {
	for i := 0; i < 30; i++ {
		smallBufferPool.Put(make([]byte, 0, SmallBufferSize))
	}

	for i := 0; i < 20; i++ {
		mediumBufferPool.Put(make([]byte, 0, MediumBufferSize))
	}

	for i := 0; i < 10; i++ {
		largeBufferPool.Put(make([]byte, 0, LargeBufferSize))
	}

	for i := 0; i < 5; i++ {
		extraLargeBufferPool.Put(make([]byte, 0, ExtraLargeBufferSize))
	}
}

func init() {
	warmupPools()
	for i := 0; i < 10; i++ {
		req := &fasthttp.Request{}
		resp := &fasthttp.Response{}
		requestPool.Put(req)
		responsePool.Put(resp)
	}
}
