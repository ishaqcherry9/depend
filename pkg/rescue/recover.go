package rescue

import (
	"context"
	"runtime/debug"

	"github.com/ishaqcherry9/depend/pkg/logger"
	"go.uber.org/zap"
)

func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logger.Error(nil, "panic recovered",
			zap.Any("err", p),
			zap.ByteString("stack", debug.Stack()),
		)
	}
}

func RecoverCtx(ctx context.Context, cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logger.Errorf(nil, "%+v\n%s", p, debug.Stack())
	}
}
