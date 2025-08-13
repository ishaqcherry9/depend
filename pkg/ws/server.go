package ws

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type ServerOption func(*serverOptions)

type serverOptions struct {
	responseHeader      http.Header
	upgrader            *websocket.Upgrader
	noClientPingTimeout time.Duration
	zapLogger           *zap.Logger
}

func defaultServerOptions() *serverOptions {
	return &serverOptions{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (o *serverOptions) apply(opts ...ServerOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithResponseHeader(header http.Header) ServerOption {
	return func(o *serverOptions) {
		o.responseHeader = header
	}
}

func WithUpgrader(upgrader *websocket.Upgrader) ServerOption {
	return func(o *serverOptions) {
		o.upgrader = upgrader
	}
}

func WithNoClientPingTimeout(timeout time.Duration) ServerOption {
	return func(o *serverOptions) {
		o.noClientPingTimeout = timeout
	}
}

func WithServerLogger(l *zap.Logger) ServerOption {
	return func(o *serverOptions) {
		if l != nil {
			o.zapLogger = l
		}
	}
}

type Conn = websocket.Conn

type LoopFn func(ctx context.Context, conn *Conn)

type Server struct {
	upgrader *websocket.Upgrader

	w              http.ResponseWriter
	r              *http.Request
	responseHeader http.Header

	noClientPingTimeout time.Duration

	loopFn LoopFn

	zapLogger *zap.Logger
}

func NewServer(w http.ResponseWriter, r *http.Request, loopFn LoopFn, opts ...ServerOption) *Server {
	o := defaultServerOptions()
	o.apply(opts...)
	if o.zapLogger == nil {
		o.zapLogger, _ = zap.NewProduction()
	}

	return &Server{
		w:      w,
		r:      r,
		loopFn: loopFn,

		upgrader:            o.upgrader,
		responseHeader:      o.responseHeader,
		noClientPingTimeout: o.noClientPingTimeout,
		zapLogger:           o.zapLogger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	conn, err := s.upgrader.Upgrade(s.w, s.r, s.responseHeader)
	if err != nil {
		return err
	}
	defer conn.Close()

	fields := []zap.Field{zap.String("client", conn.RemoteAddr().String())}
	if s.noClientPingTimeout > 0 {
		if err = conn.SetReadDeadline(time.Now().Add(s.noClientPingTimeout)); err != nil {
			return err
		}

		conn.SetPingHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(s.noClientPingTimeout))
		})
		fields = append(fields, zap.String("no_ping_timeout", fmt.Sprintf("%vs", s.noClientPingTimeout.Seconds())))
	}

	s.zapLogger.Info("new websocket connection established", fields...)

	s.loopFn(ctx, conn)

	return nil
}

func IsClientClose(err error) bool {
	return strings.Contains(err.Error(), "websocket: close") ||
		strings.Contains(err.Error(), "closed by the remote host")
}
