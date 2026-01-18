package config

type Config struct {
	Server ServerConfig `yaml:"server"`
	OIDC   OIDCConfig   `yaml:"oidc"`
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
	OIDC: OIDCConfig{
		Issuer:   "https://example.com/",
		ClientID: "your-client-id",
		Secret:   "your-client-secret",
		Scopes:   []string{"openid", "profile", "email"},
	},
	Debug: false,
}
