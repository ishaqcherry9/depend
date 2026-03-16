package middleware

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/phpserialize"
	"github.com/forgoer/openssl"
	"github.com/gin-gonic/gin"
	"github.com/ishaqcherry9/depend/pkg/logger"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// PhpSession PhpSession 中间件。
func PhpSession(cfg PhpSessionConfig, whitelistPaths []string) gin.HandlerFunc {
	if err := cfg.validate(); err != nil {
		panic(fmt.Sprintf("invalid PhpSessionConfig: %v", err))
	}

	return func(c *gin.Context) {
		// 如果路径在白名单中，则跳过验证。
		if isPathWhitelisted(c.Request.URL.Path, whitelistPaths) {
			c.Next()
			return
		}

		// 取请求参数
		tokenEncrypt := c.Request.Header.Get("Token")
		timestamp := c.Request.Header.Get("Timestamp")
		device := c.Request.Header.Get("Device")
		Customeruid := c.Request.Header.Get("Customeruid")

		if tokenEncrypt == "" || timestamp == "" || device == "" || Customeruid == "" {
			logger.Warn(c, "非法请求，缺少参数", zap.Any("header", c.Request.Header))
			c.JSON(http.StatusOK, gin.H{"code": 403, "msg": "非法请求！"})
			c.Abort()
			return
		}

		memberID, err := strconv.ParseInt(Customeruid, 10, 64)
		if err != nil {
			logger.Warn(c, "非法用户ID", zap.String("Customeruid", Customeruid), zap.Error(err))
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "非法用户ID"})
			c.Abort()
			return
		}

		// 解密 token
		token, err := DecryptToken(c, tokenEncrypt, timestamp, device, cfg.TokenConfig)
		if err != nil {
			logger.Error(c, "Token解密失败", zap.String("tokenEncrypt", tokenEncrypt), zap.Error(err))
			c.JSON(http.StatusOK, gin.H{"code": 403, "msg": "非法请求！"})
			c.Abort()
			return
		}

		// 校验 Redis 中是否存在该 token
		storedToken, err := cfg.RedisClient.HGet(context.Background(), cfg.MemberLoginTokenHashKey, fmt.Sprintf("%d", memberID)).Result()
		if err != nil || storedToken != token {
			logger.Warn(c, "登录已过期，Redis token校验失败", zap.Int64("memberID", memberID), zap.Error(err), zap.String("storedToken", storedToken), zap.String("token", token))
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "1登录已过期，请重新登录!"})
			c.Abort()
			return
		}

		// fmt.Println(storedToken)

		// 获取 session 信息
		sessionKey := cfg.SiteID + "_api_session:" + token
		sessionStr, err := cfg.RedisClient.Get(context.Background(), sessionKey).Result()
		if err != nil {
			logger.Warn(c, "登录已过期，session不存在", zap.String("sessionKey", sessionKey), zap.Error(err))
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "2登录已过期，请重新登录!"})
			c.Abort()
			return
		}

		// fmt.Println(sessionStr)

		memberInfo, parseErr := parseMemberInfo(sessionStr)
		if parseErr != nil {
			logger.Warn(c, "登录已过期，session解析失败", zap.String("session", sessionStr), zap.Error(parseErr))
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "3登录已过期，请重新登录!"})
			c.Abort()
			return
		}

		if memberInfo.MemberType != 2 && memberInfo.MemberID != memberID {
			logger.Warn(c, "登录已过期，member信息校验失败", zap.Any("memberInfo", memberInfo), zap.Int64("requestMemberID", memberID))
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "4登录已过期，请重新登录!"})
			c.Abort()
			return
		}

		// 放入 context
		c.Set("member_id", memberInfo.MemberID)
		c.Set("member_type", memberInfo.MemberType)
		c.Set("member_identity", memberInfo.MemberIdentity)
		c.Request.Header.Set("Uid", strconv.FormatInt(memberInfo.MemberID, 10))

		c.Next()
	}
}

func DecryptToken(ctx context.Context, token, timestamp, device string, config PhpSessionTokenConfig) (string, error) {

	// global.GVA_LOG.Info("收到解密Token请求", zap.String("token", token), zap.String("timestamp", timestamp), zap.String("device", device))

	// 参数校验
	if token == "" || timestamp == "" || len(timestamp) != 16 {
		logger.Warn(ctx, "参数错误", zap.String("token", token), zap.String("timestamp", timestamp), zap.String("device", device))
		return "", fmt.Errorf("非法请求，参数错误")
	}

	// 正则解析 timestamp
	reg := regexp.MustCompile(`(\d{10})(\d{3})(\d{3})`)
	match := reg.FindStringSubmatch(timestamp)
	if len(match) < 4 {
		logger.Warn(ctx, "时间戳解析失败", zap.String("timestamp", timestamp))
		return "", fmt.Errorf("非法请求，时间戳错误")
	}

	timestampInt, _ := strconv.ParseInt(match[1], 10, 64)
	part2, part3 := match[2], match[3]

	// 获取 device 对应配置
	tokenSecretKey, ok := config.KeyVal[device]
	if !ok {
		logger.Warn(ctx, "device无效", zap.String("device", device))
		return "", fmt.Errorf("非法请求，device无效")
	}

	var offset int
	var secretList []string
	var randStr string
	if tokenSecretKey == "app" {
		offset = config.App.Offset
		secretList = config.App.Secret
		randStr = config.App.RandStr
	} else if tokenSecretKey == "pc" {
		offset = config.Pc.Offset
		secretList = config.Pc.Secret
	} else {
		logger.Warn(ctx, "未知配置类型", zap.String("tokenSecretKey", tokenSecretKey))
		return "", fmt.Errorf("非法请求，未知配置")
	}

	// 计算 index
	hour := time.Unix(timestampInt, 0).Hour()
	index := (hour % 12) + offset
	if index >= len(secretList) {
		index = index % len(secretList) // 防止越界
	}

	// 生成 md5Str
	var md5Str string
	if tokenSecretKey == "app" {
		md5Str = fmt.Sprintf("%s-#%s.%s%d*%d%s%s%d",
			randStr, part3, secretList[index], offset,
			timestampInt, part2, part3, offset)
	} else {
		md5Str = fmt.Sprintf("%d==%s%s-%s,%d%s%s+%d",
			offset, part3, secretList[index], part3, timestampInt, part2, part3, offset)
	}

	// global.GVA_LOG.Debug("解密前MD5字符串", zap.String("md5Str", md5Str))

	md5Sum := md5.Sum([]byte(md5Str))
	md5Hex := hex.EncodeToString(md5Sum[:])
	key := md5Hex[offset : offset+16]

	// base64 decode
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		logger.Error(ctx, "base64解码失败", zap.Error(err))
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	// AES 解密
	plain, err := openssl.AesECBDecrypt(decoded, []byte(key), openssl.PKCS7_PADDING)
	if err != nil {
		logger.Error(ctx, "AES解密失败", zap.Error(err))
		return "", fmt.Errorf("解密失败: %v", err)
	}

	// global.GVA_LOG.Info("解密成功", zap.String("result", string(plain)))
	return string(plain), nil
}

func parseMemberInfo(sessionStr string) (*MemberInfo, error) {
	info := &MemberInfo{}

	// 去掉最外层 s:xxx:"..." 包裹
	start := strings.Index(sessionStr, `"`)
	end := strings.LastIndex(sessionStr, `"`)
	if start == -1 || end == -1 || start >= end {
		return nil, errors.New("session格式错误")
	}
	raw := sessionStr[start+1 : end]

	// 准备接收结构
	var sessionData map[interface{}]interface{}

	// 反序列化
	err := phpserialize.Unmarshal([]byte(raw), &sessionData)
	if err != nil {
		return nil, fmt.Errorf("session反序列化失败: %v", err)
	}

	// 取 member 部分
	memberVal, exists := sessionData["member"]
	if !exists {
		return nil, errors.New("member信息不存在")
	}

	memberMap, ok := memberVal.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("member信息解析失败")
	}

	if v, ok := memberMap["memberid"]; ok {
		info.MemberID = toInt64(v)
	}
	if v, ok := memberMap["member_type"]; ok {
		info.MemberType = toInt(v)
	}
	if v, ok := memberMap["member_identity"]; ok {
		info.MemberIdentity = toInt(v)
	}

	return info, nil
}

func toInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

func toInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	default:
		return 0
	}
}

// isPathWhitelisted 判断路径是否在白名单中。
func isPathWhitelisted(path string, whitelistPaths []string) bool {
	for _, item := range whitelistPaths {
		if item == "" {
			continue
		}
		if path == item {
			return true
		}
		if strings.HasSuffix(item, "*") {
			prefix := strings.TrimSuffix(item, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}
	return false
}

func EncryptToken(tokenData, timestamp, device string, config PhpSessionTokenConfig) (string, error) {
	// 参数校验
	if tokenData == "" || timestamp == "" || len(timestamp) != 16 {
		return "", fmt.Errorf("非法请求，参数错误")
	}

	// 解析 timestamp
	reg := regexp.MustCompile(`(\d{10})(\d{3})(\d{3})`)
	match := reg.FindStringSubmatch(timestamp)
	if len(match) < 4 {
		return "", fmt.Errorf("非法请求，时间戳错误")
	}
	timestampInt, _ := strconv.ParseInt(match[1], 10, 64)
	part2, part3 := match[2], match[3]

	// 获取配置
	tokenSecretKey, ok := config.KeyVal[device]
	if !ok {
		return "", fmt.Errorf("非法请求，device无效")
	}

	var offset int
	var secretList []string
	var randStr string
	if tokenSecretKey == "app" {
		offset = config.App.Offset
		secretList = config.App.Secret
		randStr = config.App.RandStr
	} else if tokenSecretKey == "pc" {
		offset = config.Pc.Offset
		secretList = config.Pc.Secret
	} else {
		return "", fmt.Errorf("非法请求，未知配置")
	}

	// 生成 index
	hour := time.Unix(timestampInt, 0).Hour()
	index := (hour % 12) + offset
	if index >= len(secretList) {
		index = index % len(secretList)
	}

	// 构造 md5Str
	var md5Str string
	if tokenSecretKey == "app" {
		md5Str = fmt.Sprintf("%s-#%s.%s%d*%d%s%s%d",
			randStr, part3, secretList[index], offset,
			timestampInt, part2, part3, offset)
	} else {
		md5Str = fmt.Sprintf("%d==%s%s-%s,%d%s%s+%d",
			offset, part3, secretList[index], part3,
			timestampInt, part2, part3, offset)
	}

	md5Sum := md5.Sum([]byte(md5Str))
	md5Hex := hex.EncodeToString(md5Sum[:])
	key := md5Hex[offset : offset+16]

	// AES 加密
	ciphertext, err := openssl.AesECBEncrypt([]byte(tokenData), []byte(key), openssl.PKCS7_PADDING)
	if err != nil {
		return "", fmt.Errorf("加密失败: %v", err)
	}

	// Base64 encode
	token := base64.StdEncoding.EncodeToString(ciphertext)
	return token, nil
}

// MemberInfo 结构
type MemberInfo struct {
	MemberID       int64
	MemberType     int
	MemberIdentity int
}

// PhpSessionTokenDeviceConfig token 设备配置。
type PhpSessionTokenDeviceConfig struct {
	Offset  int
	Secret  []string
	RandStr string
}

// PhpSessionTokenConfig token 解密配置。
type PhpSessionTokenConfig struct {
	KeyVal map[string]string
	App    PhpSessionTokenDeviceConfig
	Pc     PhpSessionTokenDeviceConfig
}

// PhpSessionConfig PhpSession 中间件依赖配置。
type PhpSessionConfig struct {
	RedisClient             *redis.Client
	TokenConfig             PhpSessionTokenConfig
	SiteID                  string
	MemberLoginTokenHashKey string
}

func (c PhpSessionConfig) validate() error {
	if c.RedisClient == nil {
		return errors.New("redis client is nil")
	}
	if c.SiteID == "" {
		return errors.New("site id is empty")
	}
	if c.MemberLoginTokenHashKey == "" {
		return errors.New("member login token hash key is empty")
	}
	if len(c.TokenConfig.KeyVal) == 0 {
		return errors.New("token key map is empty")
	}
	return nil
}
