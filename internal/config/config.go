package config

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port              string `mapstructure:"PORT"`
	Env               string `mapstructure:"APP_ENV"` // development | production
	DBHost            string `mapstructure:"DB_HOST"`
	DBPort            string `mapstructure:"DB_PORT"`
	DBUser            string `mapstructure:"DB_USER"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	DBName            string `mapstructure:"DB_NAME"`
	DBSSLMode         string `mapstructure:"DB_SSLMODE"`
	JWTSecret         string `mapstructure:"JWT_SECRET"`
	MidtransServerKey string `mapstructure:"MIDTRANS_SERVER_KEY"`
	MidtransClientKey string `mapstructure:"MIDTRANS_CLIENT_KEY"`
	MidtransIsProd    bool   `mapstructure:"MIDTRANS_IS_PROD"`
	UploadDir         string `mapstructure:"UPLOAD_DIR"`
	BaseURL           string `mapstructure:"BASE_URL"`
	CustomerURL       string `mapstructure:"CUSTOMER_URL"`
	AdminURL          string `mapstructure:"ADMIN_URL"`
	GeocoderURL       string `mapstructure:"GEOCODER_URL"`
	GeocoderEmail     string `mapstructure:"GEOCODER_EMAIL"`
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.Env, "production")
}

func Load() (*Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_NAME", "maremereso_olga")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("UPLOAD_DIR", "./uploads")
	viper.SetDefault("BASE_URL", "http://localhost:8080")
	viper.SetDefault("CUSTOMER_URL", "http://localhost:5173")
	viper.SetDefault("ADMIN_URL", "http://localhost:5174")
	// OpenStreetMap Nominatim by default. Override with a commercial geocoder in
	// production if you exceed its fair-use policy (1 req/s).
	viper.SetDefault("GEOCODER_URL", "https://nominatim.openstreetmap.org")
	viper.SetDefault("GEOCODER_EMAIL", "")
	viper.SetDefault("MIDTRANS_IS_PROD", false)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// AutomaticEnv only resolves keys Viper already knows about, and Unmarshal
	// ignores the rest. Credentials deliberately have no defaults, so each key
	// must be bound explicitly or it would be invisible when supplied through
	// the environment rather than a .env file.
	for _, key := range []string{
		"PORT", "APP_ENV",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"JWT_SECRET",
		"MIDTRANS_SERVER_KEY", "MIDTRANS_CLIENT_KEY", "MIDTRANS_IS_PROD",
		"UPLOAD_DIR", "BASE_URL", "CUSTOMER_URL", "ADMIN_URL",
		"GEOCODER_URL", "GEOCODER_EMAIL",
	} {
		if err := viper.BindEnv(key); err != nil {
			return nil, err
		}
	}

	// Optionally read .env file
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Credentials have no defaults on purpose: a missing secret must fail loudly
	// at boot rather than silently sign tokens with a well-known value.
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, errors.New("JWT_SECRET is required (set it in .env or the environment)")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if strings.TrimSpace(cfg.DBUser) == "" {
		return nil, errors.New("DB_USER is required")
	}
	if strings.TrimSpace(cfg.MidtransServerKey) == "" {
		return nil, errors.New("MIDTRANS_SERVER_KEY is required")
	}

	return &cfg, nil
}
