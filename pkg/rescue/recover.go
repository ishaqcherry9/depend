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
		logger.Error(context.Background(), "panic recovered",
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
		logger.Errorf(ctx, "%+v\n%s", p, debug.Stack())
	}
}
