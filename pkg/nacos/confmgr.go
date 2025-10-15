package nacos

import (
	"context"
	"sync"

	"github.com/ishaqcherry9/depend/pkg/conf"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go.uber.org/zap"
)

var (
	nacosOnce    sync.Once
	configClient config_client.IConfigClient
)

type NacosConf struct {
	IPAddr      string   `yaml:"ipAddr" json:"ipAddr"`
	Port        uint64   `yaml:"port" json:"port"`
	GrpcPort    uint64   `yaml:"grpcPort,omitempty" json:"grpcPort,omitempty"`
	TimeoutMs   uint64   `yaml:"timeoutMs,omitempty" json:"timeoutMs,omitempty"`
	Group       string   `yaml:"group" json:"group"`
	DataID      string   `yaml:"dataID" json:"dataID"`
	NamespaceID string   `yaml:"namespaceID" json:"namespaceID"`
	ClusterName string   `yaml:"clusterName,omitempty" json:"clusterName,omitempty"`
	LogDir      string   `yaml:"logDir,omitempty" json:"logDir,omitempty"`
	CacheDir    string   `yaml:"cacheDir,omitempty" json:"cacheDir,omitempty"`
	LogLevel    string   `yaml:"logLevel,omitempty" json:"logLevel,omitempty"`
	Username    string   `yaml:"username,omitempty" json:"username,omitempty"`
	Password    string   `yaml:"password,omitempty" json:"password,omitempty"`
	ExtDataIDs  []string `yaml:"extDataIds,omitempty" json:"extDataIds,omitempty"`
}

func MustLoad(nacosConfigFilePath string, v interface{}, f interface{}) *NacosConf {
	var (
		err    error
		config string
	)

	var nacosConfig NacosConf

	err = conf.Parse(nacosConfigFilePath, &nacosConfig)
	if err != nil {
		logger.Fatal(context.Background(), "load nacos config failed", zap.Error(err))
	}
	err = nacosConfig.InitConfigClient()
	if err != nil {
		logger.Fatal(context.Background(), "nacos config init failed", zap.Error(err))
	}
	config, err = nacosConfig.GetConfig()
	if err != nil {
		logger.Fatal(context.Background(), "nacos get config failed", zap.Error(err))
	}
	err = conf.ParseConfigData([]byte(config), "yaml", v)
	if err != nil {
		logger.Fatal(context.Background(), "load config failed", zap.Error(err))
	}

	if nacosConfig.TimeoutMs == 0 {
		nacosConfig.TimeoutMs = 2000
	}

	return &nacosConfig
}

// LoadWithWatch 加载配置并启动监听，所有nacos操作都在此函数中完成
func LoadWithWatch(nacosConfigFilePath string, target interface{}, callback func(namespace string, group string, dataId string, data string)) *NacosConf {
	var (
		err    error
		config string
	)

	var nacosConfig NacosConf

	// 1. 如果没有提供配置文件路径，使用默认路径
	if nacosConfigFilePath == "" {
		nacosConfigFilePath = "configs/nacosconf.yaml"
		logger.Info(context.Background(), "No config file path provided, using default",
			zap.String("defaultPath", nacosConfigFilePath))
	}

	// 2. 解析nacos配置文件
	err = conf.Parse(nacosConfigFilePath, &nacosConfig)
	if err != nil {
		logger.Fatal(context.Background(), "load nacos config failed", zap.Error(err))
	}

	// 3. 初始化nacos客户端
	err = nacosConfig.InitConfigClient()
	if err != nil {
		logger.Fatal(context.Background(), "nacos config init failed", zap.Error(err))
	}

	// 4. 获取配置数据
	config, err = nacosConfig.GetConfig()
	if err != nil {
		logger.Fatal(context.Background(), "nacos get config failed", zap.Error(err))
	}

	// 5. 解析配置到目标对象
	err = conf.ParseConfigData([]byte(config), "yaml", target)
	// err = conf.ParseConfigDataWithAutoDetection([]byte(config), target)
	if err != nil {
		logger.Fatal(context.Background(), "load config failed", zap.Error(err))
	}

	// 6. 设置默认超时时间
	if nacosConfig.TimeoutMs == 0 {
		nacosConfig.TimeoutMs = 2000
	}

	// 7. 启动配置监听.有回调函数时，使用回调函数,没有回调函数时,使用默认更新全局配置对象
	defaultOnChange := func(namespace, group, dataId, data string) {
		logger.Info(context.Background(), "Configuration changed, updating global config", zap.String("namespace", namespace), zap.String("group", group), zap.String("dataId", dataId), zap.String("data", data))
		// 解析新配置并更新目标对象 - 使用自动格式检测
		err = conf.ParseConfigData([]byte(config), "yaml", target)
		// err := conf.ParseConfigDataWithAutoDetection([]byte(data), target)
		if err != nil {
			logger.Error(context.Background(), "Failed to parse new config data", zap.Error(err))
			return
		}
	}
	if callback == nil {
		callback = defaultOnChange
	}
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId:   nacosConfig.DataID,
		Group:    nacosConfig.Group,
		OnChange: callback,
	})
	if err != nil {
		logger.Fatal(context.Background(), "start nacos config watch failed", zap.Error(err))
	}

	return &nacosConfig
}

func (conf *NacosConf) InitConfigClient() (err error) {

	nacosOnce.Do(func() {
		sc := []constant.ServerConfig{
			*constant.NewServerConfig(conf.IPAddr, conf.Port, constant.WithScheme("http"), constant.WithContextPath("/nacos")),
		}

		cc := *constant.NewClientConfig(
			constant.WithNamespaceId(conf.NamespaceID),
			constant.WithTimeoutMs(conf.TimeoutMs),
			constant.WithNotLoadCacheAtStart(true),
			constant.WithUsername(conf.Username),
			constant.WithPassword(conf.Password),
			//constant.WithLogDir(conf.LogDir),
			//constant.WithCacheDir(conf.CacheDir),
			constant.WithLogLevel(conf.LogLevel),
			constant.WithUpdateCacheWhenEmpty(true),
			// constant.WithTLS(constant.TLSConfig{
			// 	Enable:    false,
			// 	Appointed: true,
			// }),
		)

		configClient, err = clients.NewConfigClient(
			vo.NacosClientParam{
				ClientConfig:  &cc,
				ServerConfigs: sc,
			},
		)

		if err != nil {
			logger.Fatal(context.Background(), "new config client failed.", zap.Error(err))
		}
	})
	return
}

func (conf *NacosConf) GetConfig() (string, error) {

	mainConfig, err := configClient.GetConfig(vo.ConfigParam{DataId: conf.DataID, Group: conf.Group})
	if err != nil {
		return "", err
	}

	if len(conf.ExtDataIDs) == 0 {
		return mainConfig, nil
	}

	var configMap = make(map[interface{}]interface{})
	mainMap, err := UnmarshalYamlToMap(mainConfig)
	if err != nil {
		return "", err
	}

	var extMap = make(map[interface{}]interface{})

	for k := range conf.ExtDataIDs {
		extConfig, errMsg := configClient.GetConfig(vo.ConfigParam{DataId: conf.ExtDataIDs[k], Group: conf.Group})
		if errMsg != nil {
			return "", errMsg
		}
		tmpExtMap, errMsg := UnmarshalYamlToMap(extConfig)
		if errMsg != nil {
			return "", errMsg
		}
		extMap = MergeMap(extMap, tmpExtMap)
	}

	configMap = MergeMap(configMap, extMap)
	configMap = MergeMap(configMap, mainMap)

	yamlString, err := MarshalObjectToYamlString(configMap)
	if err != nil {
		return "", err
	}
	return yamlString, nil
}

func (conf *NacosConf) Listen(onChange func(string, string, string, string)) error {
	return configClient.ListenConfig(vo.ConfigParam{
		DataId:   conf.DataID,
		Group:    conf.Group,
		OnChange: onChange,
	})
}
