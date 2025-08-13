package rescue

import (
	"context"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"go.uber.org/zap"
	"runtime/debug"
)

func Recover(cleanups ...func()) {
	for _, cleanup := range cleanups {
		cleanup()
	}

	if p := recover(); p != nil {
		logger.Error("panic recovered",
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
		logger.Errorf("%+v\n%s", p, debug.Stack())
	}
}
