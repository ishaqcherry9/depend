package nacos

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/ishaqcherry9/depend/pkg/logger"
	"github.com/ishaqcherry9/depend/pkg/netx"
	"github.com/ishaqcherry9/depend/pkg/proc"
	"github.com/ishaqcherry9/nacos_sdk_go/clients"
	"github.com/ishaqcherry9/nacos_sdk_go/vo"
	"go.uber.org/zap"
)

func RegisterService(opts *Options) error {
	pubListenOn := figureOutListenOn(opts.ListenOn)

	host, ports, err := net.SplitHostPort(pubListenOn)
	if err != nil {
		return fmt.Errorf("failed parsing address error: %v", err)
	}
	port, _ := strconv.ParseUint(ports, 10, 16)

	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ServerConfigs: opts.ServerConfig,
			ClientConfig:  opts.ClientConfig,
		},
	)
	if err != nil {
		log.Panic(err)
	}

	_, err = client.RegisterInstance(vo.RegisterInstanceParam{
		ServiceName: opts.ServiceName,
		Ip:          host,
		Port:        port,
		Weight:      opts.Weight,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    opts.Metadata,
		ClusterName: opts.ClusterName,
		GroupName:   opts.Group,
	})

	if err != nil {
		return err
	}

	proc.AddShutdownListener(func() {
		_, err := client.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          host,
			Port:        port,
			ServiceName: opts.ServiceName,
			Cluster:     opts.Cluster,
			GroupName:   opts.Group,
			Ephemeral:   true,
		})
		if err != nil {
			logger.Info(context.Background(), "hook deregister service error: ", zap.Error(err))
		} else {
			logger.Info(context.Background(), "hook deregistered service from nacos server.")
		}
	})

	return nil
}

func DeregisterService(opts *Options) error {
	pubListenOn := figureOutListenOn(opts.ListenOn)

	host, ports, err := net.SplitHostPort(pubListenOn)
	if err != nil {
		return fmt.Errorf("failed parsing address error: %v", err)
	}
	port, _ := strconv.ParseUint(ports, 10, 16)
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ServerConfigs: opts.ServerConfig,
			ClientConfig:  opts.ClientConfig,
		},
	)
	if err != nil {
		log.Panic(err)
	}
	_, err = client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        port,
		ServiceName: opts.ServiceName,
		Cluster:     opts.Cluster,
		GroupName:   opts.Group,
		Ephemeral:   true,
	})
	if err != nil {
		logger.Error(context.Background(), "deregister service error.", zap.Error(err))
		return err
	} else {
		logger.Info(context.Background(), "deregistered service from nacos server.")
	}
	return nil
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != allEths {
		return listenOn
	}

	ip := os.Getenv(envPodIP)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
