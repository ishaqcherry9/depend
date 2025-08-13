package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ServiceBuilder struct {
	target              *target
	resolver            *AddressResolver
	cancelFunc          context.CancelFunc
	lastSnapshot        []InstanceSnapshot
	consecutiveFailures int
}

type InstanceSnapshot struct {
	Addr    string  `json:"addr"`
	Weight  float64 `json:"weight"`
	Healthy bool    `json:"healthy"`
	Enabled bool    `json:"enabled"`
}

type NacosInstanceListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Name                     string `json:"name"`
		GroupName                string `json:"groupName"`
		Clusters                 string `json:"clusters"`
		CacheMillis              int    `json:"cacheMillis"`
		Hosts                    []Host `json:"hosts"`
		LastRefTime              int64  `json:"lastRefTime"`
		Checksum                 string `json:"checksum"`
		AllIPs                   bool   `json:"allIPs"`
		ReachProtectionThreshold bool   `json:"reachProtectionThreshold"`
		Valid                    bool   `json:"valid"`
	} `json:"data"`
}

type Host struct {
	InstanceId                string            `json:"instanceId"`
	IP                        string            `json:"ip"`
	Port                      int               `json:"port"`
	Weight                    float64           `json:"weight"`
	Healthy                   bool              `json:"healthy"`
	Enabled                   bool              `json:"enabled"`
	Ephemeral                 bool              `json:"ephemeral"`
	ClusterName               string            `json:"clusterName"`
	ServiceName               string            `json:"serviceName"`
	Metadata                  map[string]string `json:"metadata"`
	InstanceHeartBeatInterval int               `json:"instanceHeartBeatInterval"`
	InstanceHeartBeatTimeOut  int               `json:"instanceHeartBeatTimeOut"`
	IPDeleteTimeout           int               `json:"ipDeleteTimeout"`
	InstanceIdGenerator       string            `json:"instanceIdGenerator"`
}

func NewServiceBuilder(serviceName string, clusters []string, nacosConf *NacosConf) (*ServiceBuilder, error) {
	if len(serviceName) <= 0 {
		return nil, fmt.Errorf("nacos service name is empty")
	}
	if nacosConf == nil {
		return nil, fmt.Errorf("nacos conf is nil")
	}
	if nacosConf.IPAddr == "" {
		return nil, fmt.Errorf("nacos IPAddr is empty")
	}
	if nacosConf.Port <= 0 {
		return nil, fmt.Errorf("nacos Port is invalid: %d", nacosConf.Port)
	}

	if len(clusters) == 0 {
		clusters = []string{"DEFAULT"}
	}

	groupName := nacosConf.Group
	if groupName == "" {
		groupName = "DEFAULT_GROUP"
	}

	namespaceID := nacosConf.NamespaceID
	if namespaceID == "" {
		namespaceID = "public"
	}

	tgt := &target{
		Addr:        fmt.Sprintf("%s:%d", nacosConf.IPAddr, nacosConf.Port),
		Service:     serviceName,
		GroupName:   groupName,
		Clusters:    clusters,
		NamespaceID: namespaceID,
		User:        nacosConf.Username,
		Password:    nacosConf.Password,
		Timeout:     time.Duration(nacosConf.TimeoutMs) * time.Millisecond,
		LogLevel:    nacosConf.LogLevel,
		LogDir:      nacosConf.LogDir,
		CacheDir:    nacosConf.CacheDir,
	}

	resolver := NewAddressResolver(serviceName)

	builder := &ServiceBuilder{
		target:   tgt,
		resolver: resolver,
	}

	return builder, nil
}

func (sb *ServiceBuilder) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	sb.cancelFunc = cancel

	pipe := make(chan []string, 10)

	go sb.startNacosPolling(ctx, pipe)

	go PopulateServiceAddresses(ctx, sb.resolver, pipe)

	logger.Infof("[Service Builder] Started service discovery for: %s", sb.target.Service)
	return nil
}

func (sb *ServiceBuilder) Stop() {
	if sb.cancelFunc != nil {
		sb.cancelFunc()
		logger.Infof("[Service Builder] Stopped service discovery for: %s", sb.target.Service)
	}
}

func (sb *ServiceBuilder) GetResolver() *AddressResolver {
	return sb.resolver
}

func (sb *ServiceBuilder) startNacosPolling(ctx context.Context, pipe chan []string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("[Service Builder] Nacos polling panic for %s: %v", sb.target.Service, r)
		}
		close(pipe)
		logger.Infof("[Service Builder] Nacos polling stopped for: %s", sb.target.Service)
	}()

	const maxFailures = 3
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger.Infof("[Service Builder] Starting Nacos polling for service: %s", sb.target.Service)

	sb.pollNacosWithRetry(pipe, maxFailures)

	for {
		select {
		case <-ticker.C:
			sb.pollNacosWithRetry(pipe, maxFailures)
		case <-ctx.Done():
			logger.Infof("[Service Builder] Context done for %s, stopping polling", sb.target.Service)
			return
		}
	}
}

func (sb *ServiceBuilder) pollNacosWithRetry(pipe chan []string, maxFailures int) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("[Service Builder] Nacos poll panic for %s: %v", sb.target.Service, r)
			sb.consecutiveFailures++
		}
	}()

	if sb.consecutiveFailures >= maxFailures {
		retryInterval := time.Duration(sb.consecutiveFailures-maxFailures+1) * 10 * time.Second
		if retryInterval > 60*time.Second {
			retryInterval = 60 * time.Second
		}
		logger.Infof("[Service Builder] %s: Too many failures (%d), waiting %v before retry",
			sb.target.Service, sb.consecutiveFailures, retryInterval)
		time.Sleep(retryInterval)
	}

	hosts, err := sb.fetchInstancesWithRetry()
	if err != nil {
		sb.consecutiveFailures++
		logger.Errorf("[Service Builder] Nacos poll failed for %s (failures: %d): %v",
			sb.target.Service, sb.consecutiveFailures, err)
		return
	}

	if sb.consecutiveFailures > 0 {
		logger.Infof("[Service Builder] %s: Recovered after %d failures",
			sb.target.Service, sb.consecutiveFailures)
		sb.consecutiveFailures = 0
	}

	currentSnapshot := sb.createInstanceSnapshot(hosts)
	validEndpoints := sb.filterValidEndpoints(hosts)

	logger.Debugf("[Service Builder] %s: found %d total instances, %d valid endpoints",
		sb.target.Service, len(hosts), len(validEndpoints))

	if sb.isSnapshotEqual(currentSnapshot, sb.lastSnapshot) {
		logger.Debugf("[Service Builder] %s: No changes detected", sb.target.Service)
		return
	}

	sb.logInstanceChanges(sb.lastSnapshot, currentSnapshot, validEndpoints)
	sb.lastSnapshot = currentSnapshot

	select {
	case pipe <- validEndpoints:
		logger.Infof("[Service Builder] Updated %d endpoints for %s", len(validEndpoints), sb.target.Service)
	default:
		logger.Warnf("[Service Builder] Channel full, dropping update for %s", sb.target.Service)
	}
}

func (sb *ServiceBuilder) fetchInstancesWithRetry() ([]Host, error) {
	const maxRetries = 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		hosts, err := sb.fetchInstancesFromNacos()
		if err == nil {
			return hosts, nil
		}

		lastErr = err
		if i < maxRetries-1 {
			retryDelay := time.Duration(i+1) * 2 * time.Second
			logger.Infof("[Service Builder] Retry %d/%d for %s after %v: %v",
				i+1, maxRetries, sb.target.Service, retryDelay, err)
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed after %d retries, last error: %w", maxRetries, lastErr)
}

func (sb *ServiceBuilder) fetchInstancesFromNacos() ([]Host, error) {
	apiURL := sb.buildNacosAPIURL()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if sb.target.User != "" && sb.target.Password != "" {
		req.SetBasicAuth(sb.target.User, sb.target.Password)
	}

	req.Header.Set("User-Agent", "cg-nacos-resolver")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debugf("[Service Builder] Raw Nacos API response: %s", string(body))

	var response NacosInstanceListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w, body: %s", err, string(body))
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("Nacos API returned error: code=%d, message=%s", response.Code, response.Message)
	}

	logger.Debugf("[Service Builder] Parsed %d hosts from Nacos API", len(response.Data.Hosts))
	return response.Data.Hosts, nil
}

func (sb *ServiceBuilder) buildNacosAPIURL() string {
	host, ports, _ := net.SplitHostPort(sb.target.Addr)
	port, _ := strconv.ParseUint(ports, 10, 16)

	baseURL := fmt.Sprintf("http://%s:%d/nacos/v2/ns/instance/list", host, port)

	params := url.Values{}
	params.Set("serviceName", sb.target.Service)
	params.Set("groupName", sb.target.GroupName)

	if sb.target.NamespaceID != "" {
		params.Set("namespaceId", sb.target.NamespaceID)
	}

	if len(sb.target.Clusters) > 0 && sb.target.Clusters[0] != "" {
		params.Set("clusterName", sb.target.Clusters[0])
	}

	params.Set("healthyOnly", "false")

	finalURL := baseURL + "?" + params.Encode()
	logger.Debugf("[Service Builder] Nacos API URL: %s", finalURL)

	return finalURL
}

func (sb *ServiceBuilder) createInstanceSnapshot(hosts []Host) []InstanceSnapshot {
	snapshot := make([]InstanceSnapshot, len(hosts))
	for i, host := range hosts {
		snapshot[i] = InstanceSnapshot{
			Addr:    fmt.Sprintf("%s:%d", host.IP, host.Port),
			Weight:  host.Weight,
			Healthy: host.Healthy,
			Enabled: host.Enabled,
		}
	}
	return snapshot
}

func (sb *ServiceBuilder) filterValidEndpoints(hosts []Host) []string {
	var validEndpoints []string

	for _, host := range hosts {
		addr := fmt.Sprintf("%s:%d", host.IP, host.Port)

		if !host.Healthy || !host.Enabled || host.Weight < 1.0 {
			continue
		}

		validEndpoints = append(validEndpoints, addr)
		logger.Debugf("[Service Builder] %s included (weight: %.1f, instanceId: %s)",
			addr, host.Weight, host.InstanceId)
	}

	return validEndpoints
}

func (sb *ServiceBuilder) isSnapshotEqual(current, last []InstanceSnapshot) bool {
	if len(current) != len(last) {
		return false
	}

	currentMap := make(map[string]InstanceSnapshot, len(current))
	for _, instance := range current {
		currentMap[instance.Addr] = instance
	}

	lastMap := make(map[string]InstanceSnapshot, len(last))
	for _, instance := range last {
		lastMap[instance.Addr] = instance
	}

	for addr, currentInstance := range currentMap {
		lastInstance, exists := lastMap[addr]
		if !exists {
			return false
		}

		if currentInstance.Weight != lastInstance.Weight ||
			currentInstance.Healthy != lastInstance.Healthy ||
			currentInstance.Enabled != lastInstance.Enabled {
			return false
		}
	}

	for addr := range lastMap {
		if _, exists := currentMap[addr]; !exists {
			return false
		}
	}

	return true
}

func (sb *ServiceBuilder) logInstanceChanges(last, current []InstanceSnapshot, validEndpoints []string) {
	lastMap := make(map[string]InstanceSnapshot)
	for _, instance := range last {
		lastMap[instance.Addr] = instance
	}

	currentMap := make(map[string]InstanceSnapshot)
	for _, instance := range current {
		currentMap[instance.Addr] = instance
	}

	var changes []string

	for addr, currentInstance := range currentMap {
		if lastInstance, exists := lastMap[addr]; !exists {
			changes = append(changes, fmt.Sprintf("Added: %s(w:%.1f,h:%v,e:%v)",
				addr, currentInstance.Weight, currentInstance.Healthy, currentInstance.Enabled))
		} else {
			var attrChanges []string
			if currentInstance.Weight != lastInstance.Weight {
				attrChanges = append(attrChanges, fmt.Sprintf("w:%.1f→%.1f", lastInstance.Weight, currentInstance.Weight))
			}
			if currentInstance.Healthy != lastInstance.Healthy {
				attrChanges = append(attrChanges, fmt.Sprintf("h:%v→%v", lastInstance.Healthy, currentInstance.Healthy))
			}
			if currentInstance.Enabled != lastInstance.Enabled {
				attrChanges = append(attrChanges, fmt.Sprintf("e:%v→%v", lastInstance.Enabled, currentInstance.Enabled))
			}

			if len(attrChanges) > 0 {
				changes = append(changes, fmt.Sprintf("Changed: %s(%s)", addr, strings.Join(attrChanges, ",")))
			}
		}
	}

	for addr, lastInstance := range lastMap {
		if _, exists := currentMap[addr]; !exists {
			changes = append(changes, fmt.Sprintf("Removed: %s(w:%.1f,h:%v,e:%v)",
				addr, lastInstance.Weight, lastInstance.Healthy, lastInstance.Enabled))
		}
	}

	changeDetail := "No specific changes"
	if len(changes) > 0 {
		changeDetail = strings.Join(changes, "; ")
	}

	logger.Infof("[Service Builder] %s instances changed (%d→%d total, %d valid): %s",
		sb.target.Service, len(last), len(current), len(validEndpoints), changeDetail)
}
