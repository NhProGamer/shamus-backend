package config

type Config struct {
	Server   ServerConfig  `yaml:"server"`
	Discord  DiscordConfig `yaml:"bot"`
	Database MongoConfig   `yaml:"database"`
	Debug    bool          `yaml:"debug"`
}

type ServerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	PublicUrl      string `yaml:"public_url"`
	CookieStoreKey string `yaml:"cookie_store_key"`
}

type DiscordConfig struct {
	ClientID int    `yaml:"client_id"`
	Secret   string `yaml:"secret"`
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
	Discord: DiscordConfig{
		ClientID: 0,
		Secret:   "",
	},
	Database: MongoConfig{
		Host:     "127.0.0.1",
		Port:     27017,
		User:     "mongoadmin",
		Password: "mongopassword",
	},
	Debug: false,
}
