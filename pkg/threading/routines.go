package threading

import (
	"bytes"
	"context"
	"github.com/ishaqcherry9/depend/pkg/rescue"
	"runtime"
	"strconv"
)

func GoSafe(fn func()) {
	go RunSafe(fn)
}

func GoSafeCtx(ctx context.Context, fn func()) {
	go RunSafeCtx(ctx, fn)
}

func RoutineId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}

func RunSafe(fn func()) {
	defer rescue.Recover()

	fn()
}

func RunSafeCtx(ctx context.Context, fn func()) {
	defer rescue.RecoverCtx(ctx)

	fn()
}
