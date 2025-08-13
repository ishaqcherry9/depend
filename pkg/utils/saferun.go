package utils

import (
	"context"
	"fmt"
	"time"
)

func SafeRun(ctx context.Context, fn func(ctx context.Context)) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()

	fn(ctx)
}

func SafeRunWithTimeout(d time.Duration, fn func(cancel context.CancelFunc)) {
	ctx, cancel := context.WithTimeout(context.Background(), d)

	go func() {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println(e)
			}
		}()

		fn(cancel)
	}()

	for range ctx.Done() {
		return
	}
}
