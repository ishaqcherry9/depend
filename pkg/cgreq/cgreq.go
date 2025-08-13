package cgreq

import (
	"context"
	"time"
)

type CgReq struct {
	Path    string            `json:"path"`
	Body    interface{}       `json:"body"`
	Headers map[string]string `json:"headers"`
	Context context.Context   `json:"-"`
	Timeout time.Duration     `json:"timeout"`
}
