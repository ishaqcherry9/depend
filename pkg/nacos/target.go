package nacos

import (
	"github.com/ishaqcherry9/depend/pkg/mapping"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"strings"
	"time"
)

type target struct {
	Addr        string        `key:"addr,optional"`
	User        string        `key:"user,optional"`
	Grpc        string        `key:"grpc,optional"`
	Password    string        `key:"password,optional"`
	AppName     string        `key:"app_name,optional"`
	Service     string        `key:"service,optional"`
	GroupName   string        `key:"group_name,optional"`
	Clusters    []string      `key:"clusters,optional"`
	NamespaceID string        `key:"namespace_id,optional"`
	Timeout     time.Duration `key:"timeout,optional"`

	LogLevel string `key:",optional"`
	LogDir   string `key:",optional"`
	CacheDir string `key:",optional"`
}

func parseURL(u string) (target, error) {
	rawURL, err := url.Parse(u)
	if err != nil {
		return target{}, errors.Wrap(err, "Malformed URL")
	}

	if rawURL.Scheme != schemeName ||
		len(rawURL.Host) == 0 || len(strings.TrimLeft(rawURL.Path, "/")) == 0 {
		return target{},
			errors.Errorf("Malformed URL('%s'). Must be in the next format: 'nacos://[user:passwd]@host/service?param=value'", u)
	}

	var tgt target
	params := make(map[string]interface{}, len(rawURL.Query()))
	for name, value := range rawURL.Query() {
		params[name] = value[0]
	}

	err = mapping.UnmarshalKey(params, &tgt)
	if err != nil {
		return target{}, errors.Wrap(err, "Malformed URL parameters")
	}

	if tgt.NamespaceID == "" {
		tgt.NamespaceID = "public"
	}

	tgt.LogLevel = os.Getenv("NACOS_LOG_LEVEL")
	tgt.LogDir = os.Getenv("NACOS_LOG_DIR")
	tgt.CacheDir = os.Getenv("NACOS_CACHE_DIR")

	tgt.User = rawURL.User.Username()
	tgt.Password, _ = rawURL.User.Password()
	tgt.Addr = rawURL.Host
	tgt.Service = strings.TrimLeft(rawURL.Path, "/")

	return tgt, nil
}

func ParseHttpURL(u string) (target, error) {
	return parseURL(u)
}
