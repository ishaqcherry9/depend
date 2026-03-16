package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestGetMessage(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []interface{}
		wantMsg  string
	}{
		{
			name:     "无参数返回模板本身",
			template: "plain",
			args:     nil,
			wantMsg:  "plain",
		},
		{
			name:     "有模板按Sprintf格式化",
			template: "id=%d,name=%s",
			args:     []interface{}{100, "tom"},
			wantMsg:  "id=100,name=tom",
		},
		{
			name:     "空模板单字符串参数直接返回",
			template: "",
			args:     []interface{}{"hello"},
			wantMsg:  "hello",
		},
		{
			name:     "空模板多参数按Sprint拼接",
			template: "",
			args:     []interface{}{"hello", 7},
			wantMsg:  "hello7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMsg, gotFields := getMessage(context.Background(), tt.template, tt.args...)
			if gotMsg != tt.wantMsg {
				t.Fatalf("消息不符合预期，want=%q, got=%q", tt.wantMsg, gotMsg)
			}
			if len(gotFields) != 0 {
				t.Fatalf("字段数量不符合预期，want=0, got=%d", len(gotFields))
			}
		})
	}
}

func TestCtxRequestIDAndAddCommonFields(t *testing.T) {
	const requestID = "trace_abc123"

	ctx := context.WithValue(context.Background(), ContextRequestIDKey, requestID)
	got := CtxRequestID(ctx)
	if got != requestID {
		t.Fatalf("CtxRequestID 不符合预期，want=%q, got=%q", requestID, got)
	}

	fields := AddCommonFields(ctx)
	if len(fields) != 1 {
		t.Fatalf("AddCommonFields 字段数量不符合预期，want=1, got=%d", len(fields))
	}
	if fields[0].Key != ContextRequestIDKey {
		t.Fatalf("AddCommonFields 字段键不符合预期，want=%q, got=%q", ContextRequestIDKey, fields[0].Key)
	}
}

func TestInfoAndInfofAndSync_Succeeds(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextRequestIDKey, "trace_test")
	payload := map[string]string{
		"test1": "test1",
		"test2": "test2",
	}

	Info(ctx, "info log", zap.Any("payload", payload))
	Infof(ctx, "payload=%+v", payload)

	if err := Sync(); err != nil {
		t.Fatalf("Sync 返回错误: %v", err)
	}
}
