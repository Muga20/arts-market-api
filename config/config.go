package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost string
	Port       string

	DBUser     string
	DBPassword string
	DBAddress  string
	DBName     string

	JWTSecret              string
	JWTExpirationInSeconds int64

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string

	ClientURL string

	// Cloudinary configuration
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

var Envs = LoadConfig()

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		return Config{}
	}

	return Config{
		PublicHost: getEnv("PUBLIC_HOST", ""),
		Port:       getEnv("PORT", ""),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBAddress:  getEnv("DB_HOST", "localhost"),
		DBName:     getEnv("DB_NAME", ""),

		JWTSecret:              getEnv("JWT_SECRET", ""),
		JWTExpirationInSeconds: getEnvAsInt("JWT_EXPIRATION_IN_SECONDS", 3600*24*7),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", ""),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),

		ClientURL: getEnv("CLIENT_URL", ""),

		// Load Cloudinary credentials
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
	}
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return intValue
}
