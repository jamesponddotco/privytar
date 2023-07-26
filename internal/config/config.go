// Package config implements the configuration logic for the Privytar service.
package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"git.sr.ht/~jamesponddotco/privytar/internal/meta"
	"git.sr.ht/~jamesponddotco/privytar/internal/timeutil"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrInvalidConfigFile is returned when the configuration file is invalid.
	ErrInvalidConfigFile xerrors.Error = "invalid configuration file"

	// ErrMissingContact is returned when the contact information is missing.
	ErrMissingContact xerrors.Error = "service's contact information is missing"

	// ErrMissingPrivacyPolicy is returned when the privacy policy is missing.
	ErrMissingPrivacyPolicy xerrors.Error = "service's privacy policy is missing"

	// ErrMissingTermsOfService is returned when the terms of service is missing.
	ErrMissingTermsOfService xerrors.Error = "service's terms of service is missing"

	// ErrMissingTLSCertificate is returned when the TLS certificate is missing.
	ErrMissingTLSCertificate xerrors.Error = "server's TLS certificate is missing"

	// ErrMissingTLSKey is returned when the TLS key is missing.
	ErrMissingTLSKey xerrors.Error = "server's TLS key is missing"

	// ErrInvalidTLSVersion is returned when the TLS version is invalid.
	ErrInvalidTLSVersion xerrors.Error = "server's TLS version is invalid; must be 1.2 or 1.3"

	// ErrInvalidHomepage is returned when the homepage is invalid.
	ErrInvalidHomepage xerrors.Error = "service's homepage is invalid"

	// ErrInvalidPrivacyPolicy is returned when the privacy policy is invalid.
	ErrInvalidPrivacyPolicy xerrors.Error = "service's privacy policy is invalid"

	// ErrInvalidTermsOfService is returned when the terms of service is invalid.
	ErrInvalidTermsOfService xerrors.Error = "service's terms of service is invalid"
)

const (
	// DefaultMinTLSVersion is the default minimum TLS version supported by the
	// server.
	DefaultMinTLSVersion string = "1.3"

	// DefaultAddress is the default address of the application.
	DefaultAddress string = ":1997"

	// DefaultPID is the default path to the PID file.
	DefaultPID string = "/var/run/privytar.pid"

	// DefaultCacheCapacity is the default capacity of the cache.
	DefaultCacheCapacity uint = 8192

	// DefaultServiceName is the default name of the service.
	DefaultServiceName string = meta.Name

	// DefaultHomepage is the default link to the service's homepage.
	DefaultHomepage string = meta.Homepage
)

// TLS represents the TLS configuration.
type TLS struct {
	// Certificate is the path to the TLS certificate.
	Certificate string `json:"certificate"`

	// Key is the path to the TLS key.
	Key string `json:"key"`

	// Version is the TLS version to use.
	Version string `json:"version"`
}

// Server represents the server configuration.
type Server struct {
	// TLS is the TLS configuration.
	TLS *TLS `json:"tls"`

	// Address is the address of the application.
	Address string `json:"address"`

	// PID is the path to the PID file.
	PID string `json:"pid"`

	// CacheCapacity is the capacity of the cache.
	CacheCapacity uint `json:"cacheCapacity"`

	// CacheTTL is the TTL of the cache.
	CacheTTL timeutil.CacheDuration `json:"cacheTTL"`
}

// Service represents the service configuration.
type Service struct {
	// Name is the name of the service.
	Name string `json:"name"`

	// Homepage is the link to the service's homepage.
	Homepage string `json:"homepage"`

	// Contact is the contact email for the service.
	Contact string `json:"contact"`

	// PrivacyPolicy is the link to the service's privacy policy.
	PrivacyPolicy string `json:"privacyPolicy"`

	// TermsOfService is the link to the service's terms of service.
	TermsOfService string `json:"termsOfService"`
}

// Config represents the application configuration.
type Config struct {
	// Service is the service configuration.
	Service *Service `json:"service"`

	// Server is the server configuration.
	Server *Server `json:"server"`
}

// LoadConfig opens a file and reads the configuration from it.
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfigFile, err)
	}
	defer file.Close()

	var cfg *Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfigFile, err)
	}

	if cfg.Server == nil {
		cfg.Server = &Server{}
	}

	if cfg.Server.TLS == nil {
		cfg.Server.TLS = &TLS{}
	}

	if cfg.Server.TLS.Version == "" {
		cfg.Server.TLS.Version = DefaultMinTLSVersion
	}

	if cfg.Server.Address == "" {
		cfg.Server.Address = DefaultAddress
	}

	if cfg.Server.PID == "" {
		cfg.Server.PID = DefaultPID
	}

	if cfg.Server.CacheCapacity == 0 {
		cfg.Server.CacheCapacity = DefaultCacheCapacity
	}

	if cfg.Server.CacheTTL.Duration == 0 {
		defaultCacheTTL := timeutil.CacheDuration{
			Duration: 60 * time.Minute,
		}

		cfg.Server.CacheTTL = defaultCacheTTL
	}

	if cfg.Service == nil {
		cfg.Service = &Service{}
	}

	if cfg.Service.Name == "" {
		cfg.Service.Name = DefaultServiceName
	}

	if cfg.Service.Homepage == "" {
		cfg.Service.Homepage = DefaultHomepage
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidConfigFile, err)
	}

	return cfg, nil
}

// Validate checks the configuration for errors.
func (cfg *Config) Validate() error {
	if cfg.Service.Contact == "" {
		return fmt.Errorf("%w", ErrMissingContact)
	}

	if cfg.Service.PrivacyPolicy == "" {
		return fmt.Errorf("%w", ErrMissingPrivacyPolicy)
	}

	if cfg.Service.TermsOfService == "" {
		return fmt.Errorf("%w", ErrMissingTermsOfService)
	}

	if cfg.Server.TLS.Certificate == "" {
		return fmt.Errorf("%w", ErrMissingTLSCertificate)
	}

	if cfg.Server.TLS.Key == "" {
		return fmt.Errorf("%w", ErrMissingTLSKey)
	}

	if cfg.Server.TLS.Version != "1.3" && cfg.Server.TLS.Version != "1.2" {
		return fmt.Errorf("%w", ErrInvalidTLSVersion)
	}

	if _, err := url.Parse(cfg.Service.Homepage); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidHomepage, err)
	}

	if _, err := url.Parse(cfg.Service.PrivacyPolicy); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidPrivacyPolicy, err)
	}

	if _, err := url.Parse(cfg.Service.TermsOfService); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTermsOfService, err)
	}

	return nil
}
