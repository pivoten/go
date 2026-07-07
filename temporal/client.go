package temporal

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"
)

// ClientConfig holds the inputs for creating a Temporal client.
type ClientConfig struct {
	Address       string
	Namespace     string
	TLS           *tls.Config             // nil = plaintext
	DataConverter converter.DataConverter // nil = SDK default
	Logger        *zap.Logger
}

// NewClient dials Temporal and verifies the connection with a health check.
func NewClient(ctx context.Context, cfg ClientConfig) (client.Client, error) {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	opts := client.Options{
		HostPort:  cfg.Address,
		Namespace: cfg.Namespace,
		Logger:    NewZapAdapter(cfg.Logger),
	}
	if cfg.TLS != nil {
		opts.ConnectionOptions = client.ConnectionOptions{TLS: cfg.TLS}
		cfg.Logger.Info("temporal: using mTLS")
	}
	if cfg.DataConverter != nil {
		opts.DataConverter = cfg.DataConverter
	}

	cfg.Logger.Info("temporal: connecting",
		zap.String("address", cfg.Address), zap.String("namespace", cfg.Namespace))

	c, err := client.Dial(opts)
	if err != nil {
		return nil, fmt.Errorf("dial temporal: %w", err)
	}
	if _, err := c.CheckHealth(ctx, &client.CheckHealthRequest{}); err != nil {
		c.Close()
		return nil, fmt.Errorf("temporal health check: %w", err)
	}
	cfg.Logger.Info("temporal: connected")
	return c, nil
}

// ConfigToClient builds a client straight from a Config (TLS + data converter
// resolved from env). Convenience for the common path.
func ConfigToClient(ctx context.Context, cfg Config, logger *zap.Logger) (client.Client, error) {
	tlsCfg, err := cfg.TLSConfig()
	if err != nil {
		return nil, err
	}
	dc, err := cfg.DataConverter()
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, ClientConfig{
		Address:       cfg.Address,
		Namespace:     cfg.Namespace,
		TLS:           tlsCfg,
		DataConverter: dc,
		Logger:        logger,
	})
}

// NewClientWithRetry retries NewClient with linear backoff, respecting ctx.
func NewClientWithRetry(ctx context.Context, cfg ClientConfig, attempts int, backoff time.Duration) (client.Client, error) {
	var last error
	for i := 0; i < attempts; i++ {
		c, err := NewClient(ctx, cfg)
		if err == nil {
			return c, nil
		}
		last = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff * time.Duration(i+1)):
		}
	}
	return nil, fmt.Errorf("temporal connect failed after %d attempts: %w", attempts, last)
}
