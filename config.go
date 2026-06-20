package awssigv4

import (
	"github.com/pucora/lura/v2/config"
)

// Namespace is the key to use to store and access the custom config data.
const Namespace = "github.com/pucora/pucora-aws-sigv4"

// Config holds AWS SigV4 signing settings.
type Config struct {
	Service       string
	Region        string
	AssumeRoleARN string
	STSRegion     string
	Debug         bool
}

func configGetter(e config.ExtraConfig) (Config, bool) {
	v, ok := e[Namespace]
	if !ok {
		return Config{}, false
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return Config{}, false
	}
	cfg := Config{}
	if v, ok := tmp["service"]; ok {
		cfg.Service, _ = v.(string)
	}
	if v, ok := tmp["region"]; ok {
		cfg.Region, _ = v.(string)
	}
	if v, ok := tmp["assume_role_arn"]; ok {
		cfg.AssumeRoleARN, _ = v.(string)
	}
	if v, ok := tmp["sts_region"]; ok {
		cfg.STSRegion, _ = v.(string)
	}
	if v, ok := tmp["debug"]; ok {
		cfg.Debug, _ = v.(bool)
	}
	if cfg.Service == "" || cfg.Region == "" {
		return Config{}, false
	}
	return cfg, true
}
