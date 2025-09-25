package ctoken

const (
	CacheModeCache   = 1
	CacheModeRedis   = 2
	CacheModeFile    = 3
	CacheModeFileDat = "gtoken.dat"

	DefaultTimeout        = 10 * 24 * 60 * 60 * 1000
	DefaultCacheKey       = "CToken:"
	DefaultTokenDelimiter = ":"
	DefaultEncryptKey     = "12345678912345678912345678912345"

	KeyUserKey    = "userKey"    // 用户标识
	KeyCreateTime = "createTime" // 创建时间
	KeyRefreshNum = "refreshNum" // 刷新次数
	KeyData       = "data"       // 缓存自定义数据
	KeyToken      = "X-Token"    // token
	KeyXDeviceID  = "X-Device-Id"
	KeyXClient    = "X-Client"
)

const (
	MsgErrUserKeyEmpty  = "userKey empty"
	MsgErrTokenEmpty    = "token is empty"
	MsgErrTokenLen      = "token len error"
	MsgErrValidate      = "user validate error"
	MsgErrDataEmpty     = "cache value is nil"
	MsgErrAuthInvalid   = "invalid authentication"
	MsgErrTokenFormat   = "invalid token format"
	MsgErrPlatformEmpty = "invalid platform info"
	MsgErrDeviceIDEmpty = "invalid deviceID info"
	MsgErrClientEmpty   = "invalid client info"
)
