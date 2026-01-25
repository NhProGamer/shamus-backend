package config

import (
	"fmt"
	"net/url"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	OIDC   OIDCConfig   `yaml:"oidc"`
	Redis  RedisConfig  `yaml:"redis"`
	Logger LoggerConfig `yaml:"logger"`
	Debug  bool         `yaml:"debug"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Pretty bool   `yaml:"pretty"` // Pretty console output
}

type ServerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	PublicUrl      string `yaml:"public_url"`
	CookieStoreKey string `yaml:"cookie_store_key"`
}

type OIDCConfig struct {
	Issuer   string   `yaml:"issuer"`
	ClientID string   `yaml:"client_id"`
	Secret   string   `yaml:"secret"`
	Scopes   []string `yaml:"scopes"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// GetIssuerURL parses and returns the OIDC issuer URL
func (oidcConfig OIDCConfig) GetIssuerURL() (*url.URL, error) {
	issuerURL, err := url.Parse(oidcConfig.Issuer)
	if err != nil {
		return nil, fmt.Errorf("invalid OIDC issuer URL: %w", err)
	}
	return issuerURL, nil
}

// GetPublicURL parses and returns the server public URL
func (serverConfig ServerConfig) GetPublicURL() (*url.URL, error) {
	publicURL, err := url.Parse(serverConfig.PublicUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid server public URL: %w", err)
	}
	return publicURL, nil
}

var defaultConfig = Config{
	Server: ServerConfig{
		Host:           "127.0.0.1",
		Port:           8080,
		PublicUrl:      "https://",
		CookieStoreKey: "",
	},
	Redis: RedisConfig{
		Host:     "",
		Port:     6379,
		Password: "",
		DB:       0,
	},
	OIDC: OIDCConfig{
		Issuer:   "https://your-issuer.com/",
		ClientID: "your-client-id",
		Secret:   "your-client-secret",
		Scopes:   []string{"openid", "profile", "email"},
	},
	Logger: LoggerConfig{
		Level:  "info",
		Pretty: true,
	},
	Debug: false,
}
