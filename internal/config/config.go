package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string `mapstructure:"PORT"`
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPassword     string `mapstructure:"DB_PASSWORD"`
	DBName         string `mapstructure:"DB_NAME"`
	DBSSLMode      string `mapstructure:"DB_SSLMODE"`
	JWTSecret      string `mapstructure:"JWT_SECRET"`
	MidtransServerKey string `mapstructure:"MIDTRANS_SERVER_KEY"`
	MidtransClientKey string `mapstructure:"MIDTRANS_CLIENT_KEY"`
	MidtransIsProd    bool   `mapstructure:"MIDTRANS_IS_PROD"`
	UploadDir      string `mapstructure:"UPLOAD_DIR"`
	BaseURL        string `mapstructure:"BASE_URL"`
	CustomerURL    string `mapstructure:"CUSTOMER_URL"`
	AdminURL       string `mapstructure:"ADMIN_URL"`
}

func Load() (*Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "ericanthony")
	viper.SetDefault("DB_PASSWORD", "")
	viper.SetDefault("DB_NAME", "maremereso_olga")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_SECRET", "cafe-olga-super-secret-jwt-key-2026-production-ready")
	viper.SetDefault("MIDTRANS_SERVER_KEY", "SB-Mid-server-TEST-SANDBOX-KEY-123")
	viper.SetDefault("MIDTRANS_CLIENT_KEY", "SB-Mid-client-TEST-SANDBOX-KEY-123")
	viper.SetDefault("MIDTRANS_IS_PROD", false)
	viper.SetDefault("UPLOAD_DIR", "./uploads")
	viper.SetDefault("BASE_URL", "http://localhost:8080")
	viper.SetDefault("CUSTOMER_URL", "http://localhost:5173")
	viper.SetDefault("ADMIN_URL", "http://localhost:5174")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Optionally read .env file
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
