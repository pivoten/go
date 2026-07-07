// Package temporal is Pivoten's shared Temporal base: a client + worker
// bootstrap with mTLS and AES-256-GCM payload encryption, configured from
// environment variables. It mirrors the conventions in
// missioncontrol-temporal-workers so services can converge on one base.
package temporal

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"go.temporal.io/sdk/converter"
)

// Config is the environment-driven Temporal configuration. Certs and the codec
// key are base64-encoded so they travel cleanly as env vars / secrets.
type Config struct {
	Address   string `envconfig:"TEMPORAL_ADDRESS" default:"localhost:7233"`
	Namespace string `envconfig:"TEMPORAL_NAMESPACE" default:"default"`
	TaskQueue string `envconfig:"TEMPORAL_TASK_QUEUE"`

	// mTLS (Temporal Cloud / self-hosted with TLS)
	ClientCertBase64   string `envconfig:"TEMPORAL_CLIENT_CERT_BASE64"`
	ClientKeyBase64    string `envconfig:"TEMPORAL_CLIENT_KEY_BASE64"`
	ServerRootCABase64 string `envconfig:"TEMPORAL_SERVER_ROOT_CA_BASE64"`
	ServerName         string `envconfig:"TEMPORAL_SERVER_NAME"`
	InsecureSkipVerify bool   `envconfig:"TEMPORAL_INSECURE_SKIP_VERIFY" default:"false"`

	// APIKey is an alternative to mTLS for Temporal Cloud.
	APIKey string `envconfig:"TEMPORAL_API_KEY"`

	// Payload encryption. When CodecKeyBase64 is set (32 bytes, base64), all
	// payloads are AES-256-GCM encrypted at rest.
	CodecKeyBase64 string `envconfig:"TEMPORAL_CODEC_KEY_BASE64"`
	CodecKeyID     string `envconfig:"TEMPORAL_CODEC_KEY_ID" default:"pivoten-1"`
}

// LoadConfig reads Config from the environment (envconfig prefix, e.g. "" or "PIVOTENCTL").
func LoadConfig(prefix string) (Config, error) {
	var c Config
	if err := envconfig.Process(prefix, &c); err != nil {
		return c, err
	}
	return c, nil
}

// TLSConfig builds a *tls.Config from the base64 cert material, or nil when no
// client cert is configured (plaintext local dev).
func (c Config) TLSConfig() (*tls.Config, error) {
	if c.ClientCertBase64 == "" && c.ClientKeyBase64 == "" {
		return nil, nil
	}
	cert, err := base64.StdEncoding.DecodeString(c.ClientCertBase64)
	if err != nil {
		return nil, fmt.Errorf("decode client cert: %w", err)
	}
	key, err := base64.StdEncoding.DecodeString(c.ClientKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode client key: %w", err)
	}
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("load client keypair: %w", err)
	}

	tc := &tls.Config{
		Certificates:       []tls.Certificate{pair},
		ServerName:         c.ServerName,
		InsecureSkipVerify: c.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}
	if c.ServerRootCABase64 != "" {
		ca, err := base64.StdEncoding.DecodeString(c.ServerRootCABase64)
		if err != nil {
			return nil, fmt.Errorf("decode server root CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("server root CA is not valid PEM")
		}
		tc.RootCAs = pool
	}
	return tc, nil
}

// DataConverter returns an encrypting converter when a codec key is configured,
// otherwise the default converter.
func (c Config) DataConverter() (converter.DataConverter, error) {
	if c.CodecKeyBase64 == "" {
		return converter.GetDefaultDataConverter(), nil
	}
	key, err := base64.StdEncoding.DecodeString(c.CodecKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode codec key: %w", err)
	}
	codec, err := NewCodec(c.CodecKeyID, key)
	if err != nil {
		return nil, err
	}
	return NewEncryptionDataConverter(codec), nil
}
