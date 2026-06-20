package awssigv4

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/transport/http/client"
)

// WrapRequestExecutor signs outbound backend requests with AWS SigV4.
func WrapRequestExecutor(cfg *config.Backend, next client.HTTPRequestExecutor) client.HTTPRequestExecutor {
	sigCfg, ok := configGetter(cfg.ExtraConfig)
	if !ok {
		return next
	}
	awsCfg, err := loadAWSConfig(sigCfg)
	if err != nil {
		return func(_ context.Context, _ *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("aws sigv4: %w", err)
		}
	}
	signer := v4.NewSigner()
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		body, err := cloneBody(req)
		if err != nil {
			return nil, err
		}
		creds, err := awsCfg.Credentials.Retrieve(ctx)
		if err != nil {
			return nil, fmt.Errorf("aws sigv4 credentials: %w", err)
		}
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
		if sigCfg.Debug {
			log.Printf("[aws-sigv4] signing %s %s service=%s region=%s", req.Method, req.URL.String(), sigCfg.Service, sigCfg.Region)
		}
		if err := signer.SignHTTP(ctx, creds, req, hashPayload(body), sigCfg.Service, sigCfg.Region, time.Now()); err != nil {
			return nil, fmt.Errorf("aws sigv4 sign: %w", err)
		}
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
		return next(ctx, req)
	}
}

func loadAWSConfig(cfg Config) (aws.Config, error) {
	ctx := context.Background()
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, err
	}
	if cfg.AssumeRoleARN != "" {
		stsRegion := cfg.STSRegion
		if stsRegion == "" {
			stsRegion = cfg.Region
		}
		stsClient := sts.NewFromConfig(awsCfg, func(o *sts.Options) {
			o.Region = stsRegion
		})
		awsCfg.Credentials = stscreds.NewAssumeRoleProvider(stsClient, cfg.AssumeRoleARN)
	}
	return awsCfg, nil
}

func cloneBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(data))
	return data, nil
}

func hashPayload(body []byte) string {
	if body == nil {
		return "UNSIGNED-PAYLOAD"
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
