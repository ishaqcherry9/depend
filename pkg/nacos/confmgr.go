package nacos

import (
	"github.com/ishaqcherry9/depend/pkg/conf"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"github.com/ishaqcherry9/nacos_sdk_go/clients"
	"github.com/ishaqcherry9/nacos_sdk_go/clients/config_client"
	"github.com/ishaqcherry9/nacos_sdk_go/common/constant"
	"github.com/ishaqcherry9/nacos_sdk_go/vo"
	"go.uber.org/zap"
	"sync"
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
		logger.Fatal("load nacos config failed", zap.Error(err))
	}
	err = nacosConfig.InitConfigClient()
	if err != nil {
		logger.Fatal("nacos config init failed", zap.Error(err))
	}
	config, err = nacosConfig.GetConfig()
	if err != nil {
		logger.Fatal("nacos get config failed", zap.Error(err))
	}
	err = conf.ParseConfigData([]byte(config), "yaml", v)
	if err != nil {
		logger.Fatal("load config failed", zap.Error(err))
	}

	if nacosConfig.TimeoutMs == 0 {
		nacosConfig.TimeoutMs = 2000
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
			constant.WithTLS(constant.TLSConfig{
				Enable:    false,
				Appointed: true,
			}),
		)

		configClient, err = clients.NewConfigClient(
			vo.NacosClientParam{
				ClientConfig:  &cc,
				ServerConfigs: sc,
			},
		)

		if err != nil {
			logger.Fatal("new config client failed.", zap.Error(err))
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
