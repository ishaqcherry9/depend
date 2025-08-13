package nacos

import (
	"fmt"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"sync/atomic"
)

type ServiceClient struct {
	serviceName    string
	serviceBuilder *ServiceBuilder
	counter        atomic.Uint64
}

func NewServiceClient(serviceName string, clusters []string, nacosConf *NacosConf) (*ServiceClient, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	if nacosConf == nil {
		return nil, fmt.Errorf("nacos config cannot be nil")
	}

	builder, err := NewServiceBuilder(serviceName, clusters, nacosConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create service builder: %w", err)
	}

	err = builder.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start service discovery: %w", err)
	}

	client := &ServiceClient{
		serviceName:    serviceName,
		serviceBuilder: builder,
	}

	logger.Infof("[Service Client] Created client for service: %s", serviceName)
	return client, nil
}

func (sc *ServiceClient) GetAddresses() []string {
	return sc.serviceBuilder.GetResolver().GetAddresses()
}

func (sc *ServiceClient) GetServiceName() string {
	return sc.serviceName
}
