package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestSync_Succeeds(t *testing.T) {
	a := map[string]string{
		"test1": "test1",
		"test2": "test2",
		"test3": "test3",
		"test4": "test4",
		"test5": "test5",
	}
	Info(context.Background(), "", zap.Any("test0", a))
	Infof(context.Background(), "%+v", a)

}
