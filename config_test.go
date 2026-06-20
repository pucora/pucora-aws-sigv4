package awssigv4

import (
	"testing"

	"github.com/pucora/lura/v2/config"
)

func TestConfigGetterRequiresServiceAndRegion(t *testing.T) {
	cfg := &config.Backend{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				"service": "execute-api",
			},
		},
	}
	if _, ok := configGetter(cfg.ExtraConfig); ok {
		t.Fatal("expected missing region to disable config")
	}
}

func TestConfigGetterParsesFields(t *testing.T) {
	cfg := &config.Backend{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				"service":         "execute-api",
				"region":          "us-east-1",
				"assume_role_arn": "arn:aws:iam::123:role/Gateway",
				"sts_region":      "us-west-2",
				"debug":           true,
			},
		},
	}
	got, ok := configGetter(cfg.ExtraConfig)
	if !ok {
		t.Fatal("expected valid config")
	}
	if got.Service != "execute-api" || got.Region != "us-east-1" {
		t.Fatalf("unexpected service/region: %+v", got)
	}
	if got.AssumeRoleARN == "" || got.STSRegion != "us-west-2" || !got.Debug {
		t.Fatalf("unexpected config: %+v", got)
	}
}

func TestHashPayload(t *testing.T) {
	if hashPayload(nil) != "UNSIGNED-PAYLOAD" {
		t.Fatal("expected unsigned payload marker")
	}
	if hashPayload([]byte("hello")) == "UNSIGNED-PAYLOAD" {
		t.Fatal("expected hashed payload")
	}
}
