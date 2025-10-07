package captchabase64

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/ishaqcherry9/depend/pkg/goredis"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"go.uber.org/zap"
)

func newTestService(t *testing.T) (*base64CaptchaService, *goredis.Client, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	cleanup := func() { mr.Close() }

	cli, err := goredis.InitSingle(mr.Addr(), "", 0)
	if err != nil {
		cleanup()
		t.Fatalf("init redis: %v", err)
	}

	cfg := &Base64CaptchaServiceConfig{
		client: cli,
		// keyPrefix:   "captcha:",
		// expiration:  2 * time.Minute,
		// width:       120,
		// height:      40,
		// noiseCount:  1,
		// dotCount:    1,
		// length:      4,
		// maxSkew:     0.4,
		// charset:     "1234567890abcdef",
		// lineOptions: 2,
	}
	if _, err := NewCaptchaServiceWithConfig(context.Background(), cfg); err != nil {
		cleanup()
		t.Fatalf("init service: %v", err)
	}
	svc, err := GetBase64CaptchaService()
	if err != nil {
		cleanup()
		t.Fatalf("get service: %v", err)
	}
	return svc, cli, cleanup
}

func mustBase64(t *testing.T, s string) {
	if s == "" {
		t.Fatalf("empty base64 string")
	}
	// strip data URL prefix if present
	if idx := strings.Index(s, ","); idx != -1 && strings.HasPrefix(s, "data:") {
		s = s[idx+1:]
	}
	if _, err := base64.StdEncoding.DecodeString(s); err != nil {
		t.Fatalf("invalid base64: %v", err)
	}
}

func TestGenerateDigitAndVerify(t *testing.T) {
	svc, cli, done := newTestService(t)
	defer done()

	id, b64, answer, err := svc.GenerateDigit(context.Background())
	if err != nil {
		t.Fatalf("GenerateDigit error: %v", err)
	}
	if id == "" {
		t.Fatalf("empty id")
	}
	mustBase64(t, b64)

	// fetch expected answer from redis and verify
	answer, err = cli.Get(context.Background(), "captcha:"+id).Result()
	if err != nil {
		t.Fatalf("get answer: %v", err)
	}
	if !svc.Verify(context.Background(), id, answer, true) {
		t.Fatalf("verify failed with correct answer")
	}
	// cleared after verify
	if svc.Verify(context.Background(), id, answer, false) {
		t.Fatalf("verify should fail after clear")
	}
	// // 将b64img保存到本地文件
	// b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	// imageData, err := base64.StdEncoding.DecodeString(b64)
	// if err != nil {
	// 	t.Fatalf("base64图片解码失败: %v", err)
	// }
	// os.WriteFile("captcha.png", imageData, 0644)
	// defer func() {
	// 	time.Sleep(5 * time.Second)
	// 	os.Remove("captcha.png")
	// }()
	logger.Debug(context.Background(), "captcha info", zap.String("id", id), zap.String("answer", answer))
}

func TestGenerateDigitWith(t *testing.T) {
	svc, cli, done := newTestService(t)
	defer done()

	id, b64, answer, err := svc.GenerateDigitWith(context.Background(), 200, 60, 5, 0.5, 2)
	if err != nil {
		t.Fatalf("GenerateDigitWith error: %v", err)
	}
	if id == "" {
		t.Fatalf("empty id")
	}
	mustBase64(t, b64)

	answer, err = cli.Get(context.Background(), "captcha:"+id).Result()
	if err != nil {
		t.Fatalf("get answer: %v", err)
	}
	if !svc.Verify(context.Background(), id, answer, true) {
		t.Fatalf("verify failed with correct answer")
	}

	// // 将b64img保存到本地文件
	// b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	// imageData, err := base64.StdEncoding.DecodeString(b64)
	// if err != nil {
	// 	t.Fatalf("base64图片解码失败: %v", err)
	// }
	// os.WriteFile("captcha.png", imageData, 0644)
	// defer func() {
	// 	time.Sleep(5 * time.Second)
	// 	os.Remove("captcha.png")
	// }()
	logger.Debug(context.Background(), "captcha info", zap.String("id", id), zap.String("answer", answer))
}

func TestGenerateMathAndVerify(t *testing.T) {
	svc, _, done := newTestService(t)
	defer done()

	id, b64, answer, err := svc.GenerateMath(context.Background())
	if err != nil {
		t.Fatalf("GenerateMath error: %v", err)
	}
	if id == "" || answer == "" {
		t.Fatalf("empty id/answer")
	}
	mustBase64(t, b64)
	if !svc.Verify(context.Background(), id, answer, true) {
		t.Fatalf("verify failed with correct answer")
	}
	// // 将b64img保存到本地文件
	// b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	// imageData, err := base64.StdEncoding.DecodeString(b64)
	// if err != nil {
	// 	t.Fatalf("base64图片解码失败: %v", err)
	// }
	// os.WriteFile("captcha.png", imageData, 0644)
	// defer func() {
	// 	time.Sleep(5 * time.Second)
	// 	os.Remove("captcha.png")
	// }()
	logger.Debug(context.Background(), "captcha info", zap.String("id", id), zap.String("answer", answer))
}

func TestGenerateMathWith(t *testing.T) {
	svc, _, done := newTestService(t)
	defer done()

	id, b64, answer, err := svc.GenerateMathWith(context.Background(), 180, 60, 3)
	if err != nil {
		t.Fatalf("GenerateMathWith error: %v", err)
	}
	if id == "" || answer == "" {
		t.Fatalf("empty id/answer")
	}
	mustBase64(t, b64)
	if !svc.Verify(context.Background(), id, answer, true) {
		t.Fatalf("verify failed with correct answer")
	}
	// // 将b64img保存到本地文件
	// b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	// imageData, err := base64.StdEncoding.DecodeString(b64)
	// if err != nil {
	// 	t.Fatalf("base64图片解码失败: %v", err)
	// }
	// os.WriteFile("captcha.png", imageData, 0644)
	// defer func() {
	// 	time.Sleep(5 * time.Second)
	// 	os.Remove("captcha.png")
	// }()
	logger.Debug(context.Background(), "captcha info", zap.String("id", id), zap.String("answer", answer))
}

func TestGenerateStringAndVerify(t *testing.T) {
	svc, cli, done := newTestService(t)
	defer done()

	id, b64, answer, err := svc.GenerateString(context.Background())
	if err != nil {
		t.Fatalf("GenerateString error: %v", err)
	}
	if id == "" {
		t.Fatalf("empty id")
	}
	mustBase64(t, b64)

	answer, err = cli.Get(context.Background(), "captcha:"+id).Result()
	if err != nil {
		t.Fatalf("get answer: %v", err)
	}
	if !svc.Verify(context.Background(), id, answer, true) {
		t.Fatalf("verify failed with correct answer")
	}
	// // 将b64img保存到本地文件
	// b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	// imageData, err := base64.StdEncoding.DecodeString(b64)
	// if err != nil {
	// 	t.Fatalf("base64图片解码失败: %v", err)
	// }
	// os.WriteFile("captcha.png", imageData, 0644)
	// defer func() {
	// 	time.Sleep(5 * time.Second)
	// 	os.Remove("captcha.png")
	// }()
	logger.Debug(context.Background(), "captcha info", zap.String("id", id), zap.String("answer", answer))
}

func TestGenerateStringWith(t *testing.T) {
	svc, cli, done := newTestService(t)
	defer done()

	id, b64, answer, err := svc.GenerateStringWith(context.Background(), 160, 50, 5, 2, 4, "abcd1234", nil)
	if err != nil {
		t.Fatalf("GenerateStringWith error: %v", err)
	}
	if id == "" {
		t.Fatalf("empty id")
	}
	mustBase64(t, b64)

	answer, err = cli.Get(context.Background(), "captcha:"+id).Result()
	if err != nil {
		t.Fatalf("get answer: %v", err)
	}
	if !svc.Verify(context.Background(), id, answer, true) {
		t.Fatalf("verify failed with correct answer")
	}
	// // 将b64img保存到本地文件
	// b64 = strings.TrimPrefix(b64, "data:image/png;base64,")
	// imageData, err := base64.StdEncoding.DecodeString(b64)
	// if err != nil {
	// 	t.Fatalf("base64图片解码失败: %v", err)
	// }
	// defer func() {
	// 	time.Sleep(5 * time.Second)
	// 	os.Remove("captcha.png")
	// }()
	// os.WriteFile("captcha.png", imageData, 0644)
	logger.Debug(context.Background(), "captcha info", zap.String("id", id), zap.String("answer", answer))
}
