package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type LogConfig struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	Output    string `mapstructure:"output"`
	LogFile   string `mapstructure:"log_file"`
	ErrorFile string `mapstructure:"error_file"`
	DebugBody bool   `mapstructure:"debug_body"`
}

type JWTConfig struct {
	PrivateKeyPath   string        `mapstructure:"private_key_path"`   // RSA 私钥文件路径
	PublicKeyPath    string        `mapstructure:"public_key_path"`    // RSA 公钥文件路径
	AccessTTL        time.Duration `mapstructure:"access_ttl"`         // Access Token 有效期
	RefreshTTL       time.Duration `mapstructure:"refresh_ttl"`        // Refresh Token 有效期
	AutoGenerateKeys bool          `mapstructure:"auto_generate_keys"` // 是否自动生成密钥对
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
