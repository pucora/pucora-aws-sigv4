package awssigv4

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/pucora/lura/v2/config"
)

func TestSignHTTPWithStaticCredentials(t *testing.T) {
	creds := aws.Credentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}
	req, err := http.NewRequest(http.MethodGet, "https://example.execute-api.us-east-1.amazonaws.com/prod/hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"hello":"world"}`)
	req.Body = io.NopCloser(bytes.NewReader(body))
	signer := v4.NewSigner()
	if err := signer.SignHTTP(context.Background(), creds, req, hashPayload(body), "execute-api", "us-east-1", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Authorization") == "" {
		t.Fatal("expected Authorization header after signing")
	}
	if req.Header.Get("X-Amz-Date") == "" {
		t.Fatal("expected X-Amz-Date header after signing")
	}
}

func TestWrapRequestExecutorPassthroughWithoutConfig(t *testing.T) {
	called := false
	cfg := &config.Backend{ExtraConfig: config.ExtraConfig{}}
	next := func(_ context.Context, _ *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	}
	exec := WrapRequestExecutor(cfg, next)
	if _, err := exec(context.Background(), &http.Request{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected passthrough executor")
	}
}
