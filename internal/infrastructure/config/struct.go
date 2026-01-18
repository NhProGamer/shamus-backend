package config

import "net/url"

type Config struct {
	Server ServerConfig `yaml:"server"`
	OIDC   OIDCConfig   `yaml:"oidc"`
	Redis  RedisConfig  `yaml:"redis"`
	Debug  bool         `yaml:"debug"`
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

func (oidcConfig OIDCConfig) GetIssuerURL() url.URL {
	issuerURL, err := url.Parse(oidcConfig.Issuer)
	if err != nil {
		panic("Invalid OIDC issuer URL in configuration: " + err.Error())
	}
	return *issuerURL
}

func (serverConfig ServerConfig) GetPublicURL() url.URL {
	issuerURL, err := url.Parse(serverConfig.PublicUrl)
	if err != nil {
		panic("Invalid OIDC issuer URL in configuration: " + err.Error())
	}
	return *issuerURL
}

type MongoConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
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
	Debug: false,
}
