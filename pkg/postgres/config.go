package postgres

import (
	"fmt"
)

// Config is a minimal postgres config for small services.
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

func (c Config) normalize() Config {
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 20
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 20
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		c.MaxIdleConns = c.MaxOpenConns
	}
	return c
}

func (c Config) validate() error {
	if c.User == "" {
		return fmt.Errorf("postgres config: user is required")
	}
	if c.Database == "" {
		return fmt.Errorf("postgres config: database is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("postgres config: invalid port")
	}
	if c.MaxOpenConns < 1 {
		return fmt.Errorf("postgres config: max open conns must be > 0")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("postgres config: max idle conns must be >= 0")
	}
	return nil
}

func (c Config) dsn() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
		c.SSLMode,
	)
}
