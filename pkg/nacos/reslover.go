package nacos

import (
	"context"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"sort"
	"sync"
)

const schemeName = "nacos"

type AddressResolver struct {
	serviceName string
	addresses   []string
	mutex       sync.RWMutex
}

func NewAddressResolver(serviceName string) *AddressResolver {
	return &AddressResolver{
		serviceName: serviceName,
		addresses:   make([]string, 0),
	}
}

func (resolver *AddressResolver) updateAddresses(endpoints []string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()

	endpointSet := make(map[string]struct{}, len(endpoints))
	for _, endpoint := range endpoints {
		if endpoint != "" {
			endpointSet[endpoint] = struct{}{}
		}
	}

	addresses := make([]string, 0, len(endpointSet))
	for endpoint := range endpointSet {
		addresses = append(addresses, endpoint)
	}

	sort.Strings(addresses)

	resolver.addresses = addresses

	logger.Infof("[Address Resolver] Updated service: %s, addresses: %v", resolver.serviceName, addresses)
}

func (resolver *AddressResolver) GetAddresses() []string {
	resolver.mutex.RLock()
	defer resolver.mutex.RUnlock()

	if len(resolver.addresses) == 0 {
		return []string{}
	}

	result := make([]string, len(resolver.addresses))
	copy(result, resolver.addresses)
	return result
}

func (resolver *AddressResolver) GetServiceName() string {
	return resolver.serviceName
}

func (resolver *AddressResolver) GetAddressCount() int {
	resolver.mutex.RLock()
	defer resolver.mutex.RUnlock()

	return len(resolver.addresses)
}

func (resolver *AddressResolver) HasAddresses() bool {
	resolver.mutex.RLock()
	defer resolver.mutex.RUnlock()

	return len(resolver.addresses) > 0
}

func (resolver *AddressResolver) Clear() {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()

	resolver.addresses = make([]string, 0)
	logger.Infof("[Address Resolver] Cleared addresses for service: %s", resolver.serviceName)
}

func PopulateServiceAddresses(ctx context.Context, resolver *AddressResolver, input <-chan []string) {
	serviceName := resolver.GetServiceName()

	for {
		select {
		case endpoints := <-input:
			if len(endpoints) == 0 {
				logger.Warnf("[Address Resolver] Received empty endpoints for service: %s", serviceName)
				resolver.Clear()
				continue
			}

			endpointSet := make(map[string]struct{}, len(endpoints))
			for _, endpoint := range endpoints {
				if endpoint != "" {
					endpointSet[endpoint] = struct{}{}
				}
			}

			addresses := make([]string, 0, len(endpointSet))
			for endpoint := range endpointSet {
				addresses = append(addresses, endpoint)
				logger.Infof("[Address Resolver] Preparing address for %s: %s", serviceName, endpoint)
			}

			if len(addresses) == 0 {
				logger.Warnf("[Address Resolver] No valid addresses for service: %s", serviceName)
				resolver.Clear()
				continue
			}

			resolver.updateAddresses(addresses)

			logger.Infof("[Address Resolver] Successfully updated %d addresses for service: %s", len(addresses), serviceName)

		case <-ctx.Done():
			logger.Infof("[Address Resolver] Address watcher for service %s has been finished", serviceName)
			return
		}
	}
}
