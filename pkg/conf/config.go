package conf

type AppConfig struct {
	Name                  string                 `yaml:"name" json:"name"`
	SiteID                string                 `yaml:"siteID" json:"siteID"`
	Env                   string                 `yaml:"env" json:"env"`
	Version               string                 `yaml:"version" json:"version"`
	Host                  string                 `yaml:"host" json:"host"`
	EnableStat            bool                   `yaml:"enableStat" json:"enableStat"`
	EnableMetrics         bool                   `yaml:"enableMetrics" json:"enableMetrics"`
	EnableHTTPProfile     bool                   `yaml:"enableHTTPProfile" json:"enableHTTPProfile"`
	EnableLimit           bool                   `yaml:"enableLimit" json:"enableLimit"`
	EnableCircuitBreaker  bool                   `yaml:"enableCircuitBreaker" json:"enableCircuitBreaker"`
	EnableTrace           bool                   `yaml:"enableTrace" json:"enableTrace"`
	TracingSamplingRate   float64                `yaml:"tracingSamplingRate" json:"tracingSamplingRate"`
	CacheType             string                 `yaml:"cacheType" json:"cacheType"`
	Encrypt               bool                   `yaml:"encrypt" json:"encrypt"`
	TopicName             string                 `yaml:"topicName" json:"topicName"`
	RegistryDiscoveryType string                 `yaml:"registryDiscoveryType" json:"registryDiscoveryType"`
	Extend                map[string]interface{} `yaml:"extend" json:"extend"`
}

type JaegerConfig struct {
	AgentHost string                 `yaml:"agentHost" json:"agentHost"`
	AgentPort int                    `yaml:"agentPort" json:"agentPort"`
	Extend    map[string]interface{} `yaml:"extend" json:"extend"`
}

type MySqlConfig struct {
	Name            string                 `yaml:"name" json:"name"`
	IP              string                 `yaml:"ip" json:"ip"`
	Port            int                    `yaml:"port" json:"port"`
	DBName          string                 `yaml:"dbName" json:"dbName"`
	Username        string                 `yaml:"username" json:"username"`
	Password        string                 `yaml:"password" json:"password"`
	EnableLog       bool                   `yaml:"enableLog" json:"enableLog"`
	MaxIdleConns    int                    `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxOpenConns    int                    `yaml:"maxOpenConns" json:"maxOpenConns"`
	ConnMaxLifetime int                    `yaml:"connMaxLifetime" json:"connMaxLifetime"`
	Extend          map[string]interface{} `yaml:"extend" json:"extend"`
}

type MysqlExtConfig struct {
	MaxIdleConns    int                    `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxOpenConns    int                    `yaml:"maxOpenConns" json:"maxOpenConns"`
	ConnMaxLifetime int                    `yaml:"connMaxLifetime" json:"connMaxLifetime"`
	Extend          map[string]interface{} `yaml:"extend" json:"extend"`
}

type DorisConfig struct {
	IP        string                 `yaml:"ip" json:"ip"`
	Port      int                    `yaml:"port" json:"port"`
	DBName    string                 `yaml:"dbName" json:"dbName"`
	Username  string                 `yaml:"username" json:"username"`
	Password  string                 `yaml:"password" json:"password"`
	EnableLog bool                   `yaml:"enableLog" json:"enableLog"`
	Extend    map[string]interface{} `yaml:"extend" json:"extend"`
}

type RedisConfig struct {
	Name         string                 `yaml:"name" json:"name"`
	IP           string                 `yaml:"ip" json:"ip"`
	Port         int                    `yaml:"port" json:"port"`
	Username     string                 `yaml:"username" json:"username"`
	Password     string                 `yaml:"password" json:"password"`
	DialTimeout  int                    `yaml:"dialTimeout" json:"dialTimeout"`
	ReadTimeout  int                    `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout int                    `yaml:"writeTimeout" json:"writeTimeout"`
	Db           int                    `yaml:"db" json:"db"`
	Extend       map[string]interface{} `yaml:"extend" json:"extend"`
}

type LoggerConfig struct {
	Format        string                 `yaml:"format" json:"format"`
	IsSave        bool                   `yaml:"isSave" json:"isSave"`
	Level         string                 `yaml:"level" json:"level"`
	LogFileConfig LogFileConfig          `yaml:"logFileConfig" json:"logFileConfig"`
	IsCompression bool                   `yaml:"isCompression" json:"isCompression"`
	Extend        map[string]interface{} `yaml:"extend" json:"extend"`
}

type LogFileConfig struct {
	FileName      string                 `yaml:"filename" json:"filename"`
	MaxSize       int64                  `yaml:"maxSize" json:"maxSize"`
	MaxBackups    int64                  `yaml:"maxBackups" json:"maxBackups"`
	MaxAge        int64                  `yaml:"maxAge" json:"maxAge"`
	IsCompression bool                   `yaml:"isCompression" json:"isCompression"`
	Extend        map[string]interface{} `yaml:"extend" json:"extend"`
}

type HttpConfig struct {
	Port    int                    `yaml:"port" json:"port"`
	Timeout int                    `yaml:"timeout" json:"timeout"`
	Extend  map[string]interface{} `yaml:"extend" json:"extend"`
}

type CTokenConfig struct {
	Enable           bool                   `yaml:"enable" json:"enable"`
	CacheMode        int8                   `yaml:"cacheMode" json:"cacheMode"`
	CachePreKey      string                 `yaml:"cachePreKey" json:"cachePreKey"`
	Timeout          int64                  `yaml:"timeout" json:"timeout"`
	MaxRefresh       int64                  `yaml:"maxRefresh" json:"maxRefresh"`
	MaxRefreshTimes  int                    `yaml:"maxRefreshTimes" json:"maxRefreshTimes"`
	TokenDelimiter   string                 `yaml:"tokenDelimiter" json:"tokenDelimiter"`
	EncryptKey       string                 `yaml:"encryptKey" json:"encryptKey"`
	MultiLogin       bool                   `yaml:"multiLogin" json:"multiLogin"`
	AuthExcludePaths []string               `yaml:"authExcludePaths" json:"authExcludePaths"`
	Extend           map[string]interface{} `yaml:"extend" json:"extend"`
}
