package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/alexandrmotologa/clicklens/internal/models"
)

// Client wraps the ClickHouse driver connection and health tracking.
type Client struct {
	mu           sync.RWMutex
	conn         driver.Conn
	config       models.ConnectionConfig
	version      string
	connected    bool
	simulator    *Simulator
	isSimulated  bool
}

// NewClient initializes a ClickHouse client.
func NewClient(cfg models.ConnectionConfig) (*Client, error) {
	c := &Client{
		config:    cfg,
		simulator: NewSimulator(),
	}

	if cfg.IsDemo || cfg.DSN == "" || cfg.DSN == "demo" {
		c.isSimulated = true
		c.connected = true
		c.version = "24.8.3.16 (ClickLens Simulated Engine)"
		return c, nil
	}

	err := c.Connect(cfg.DSN)
	if err != nil {
		// Fallback to simulator if connection fails so the UI remains interactive
		c.isSimulated = true
		c.connected = true
		c.version = "24.8.3.16 (ClickLens Simulated Engine - Live Connect Failed)"
		return c, nil
	}

	return c, nil
}

// Connect parses the DSN and establishes a connection pool.
func (c *Client) Connect(dsn string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	if dsn == "" || dsn == "demo" {
		c.isSimulated = true
		c.connected = true
		c.config.IsDemo = true
		c.config.DSN = "demo"
		c.version = "24.8.3.16 (ClickLens Simulated Engine)"
		return nil
	}

	opts, err := parseDSN(dsn)
	if err != nil {
		return fmt.Errorf("invalid DSN format: %w", err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return fmt.Errorf("clickhouse ping failed: %w", err)
	}

	var version string
	row := conn.QueryRow(ctx, "SELECT version()")
	if err := row.Scan(&version); err != nil {
		version = "unknown"
	}

	c.conn = conn
	c.config.DSN = dsn
	c.config.IsDemo = false
	c.isSimulated = false
	c.connected = true
	c.version = version

	return nil
}

// parseDSN converts URL strings like clickhouse://user:pass@host:9000/db or http://... to clickhouse.Options.
func parseDSN(rawDSN string) (*clickhouse.Options, error) {
	if !strings.Contains(rawDSN, "://") {
		rawDSN = "clickhouse://" + rawDSN
	}

	u, err := url.Parse(rawDSN)
	if err != nil {
		return nil, err
	}

	scheme := strings.ToLower(u.Scheme)
	protocol := clickhouse.Native
	if scheme == "http" || scheme == "https" {
		protocol = clickhouse.HTTP
	}

	addr := u.Host
	if !strings.Contains(addr, ":") {
		if protocol == clickhouse.HTTP {
			addr += ":8123"
		} else {
			addr += ":9000"
		}
	}

	user := "default"
	password := ""
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password()
	}

	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		database = "default"
	}

	var tlsConfig *tls.Config
	if scheme == "https" || u.Query().Get("secure") == "true" {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	opts := &clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		Protocol: protocol,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		TLS: tlsConfig,
	}

	return opts, nil
}

// IsSimulated reports whether client is using simulation mode.
func (c *Client) IsSimulated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isSimulated
}

// GetStatus returns the current connection health and metadata.
func (c *Client) GetStatus(ctx context.Context) models.ServerStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := models.ServerStatus{
		Connected:   c.connected,
		Version:     c.version,
		CurrentDSN:  c.config.DSN,
		Protocol:    c.config.Protocol,
		IsDemo:      c.isSimulated,
		LastChecked: time.Now(),
	}

	if c.isSimulated {
		simStatus := c.simulator.GetStatus()
		status.UptimeSeconds = simStatus.UptimeSeconds
		status.ActiveQueries = simStatus.ActiveQueries
		status.ActiveMerges = simStatus.ActiveMerges
		status.TotalParts = simStatus.TotalParts
		status.TotalBytesDisk = simStatus.TotalBytesDisk
		return status
	}

	if c.conn != nil {
		var uptime uint64
		_ = c.conn.QueryRow(ctx, "SELECT uptime()").Scan(&uptime)
		status.UptimeSeconds = uptime

		var activeQueries int
		_ = c.conn.QueryRow(ctx, "SELECT count() FROM system.processes").Scan(&activeQueries)
		status.ActiveQueries = activeQueries

		var activeMerges int
		_ = c.conn.QueryRow(ctx, "SELECT count() FROM system.merges").Scan(&activeMerges)
		status.ActiveMerges = activeMerges

		var totalParts int
		var totalBytes uint64
		_ = c.conn.QueryRow(ctx, "SELECT count(), sum(bytes_on_disk) FROM system.parts WHERE active = 1").Scan(&totalParts, &totalBytes)
		status.TotalParts = totalParts
		status.TotalBytesDisk = totalBytes
	}

	return status
}

// Conn returns the raw ClickHouse driver connection, if connected.
func (c *Client) Conn() driver.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// Simulator returns the simulator instance.
func (c *Client) Simulator() *Simulator {
	return c.simulator
}

// Close closes the underlying connection pool.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
