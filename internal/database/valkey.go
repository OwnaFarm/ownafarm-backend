package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/valkey-io/valkey-go"
)

var Valkey valkey.Client

func ConnectValkey(cfg *config.ValkeyConfig) error {
	opt := valkey.ClientOption{
		InitAddress: []string{cfg.Addr},
		Password:    cfg.Password,
		SelectDB:    cfg.DB,
	}
	if cfg.TLS {
		opt.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client, err := valkey.NewClient(opt)
	if err != nil {
		return fmt.Errorf("failed to connect to valkey: %w", err)
	}

	// Test Connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Do(ctx, client.B().Ping().Build()).Error()
	if err != nil {
		return fmt.Errorf("failed to ping valkey: %w", err)
	}

	Valkey = client
	fmt.Println("Valkey connected")

	return nil
}

func CloseValkey() {
	if Valkey != nil {
		Valkey.Close()
	}
}
