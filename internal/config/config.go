package config

import "os"

type Config struct {
	Port       string
	DBURL      string
	QdrantURL  string
	JWTSecret  string
	Env        string
	S3Bucket   string
	S3Endpoint string
	AWSRegion  string
}

func Load() Config {
	return Config{
		Port:       getEnv("PORT", "8080"),
		DBURL:      getEnv("DB_URL", "postgres://postgres:postgres@db:5432/jlrdi?sslmode=disable"),
		QdrantURL:  getEnv("QDRANT_URL", "http://qdrant:6333"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-key-change-in-production"),
		Env:        getEnv("ENV", "dev"),
		S3Bucket:   getEnv("S3_BUCKET", "jlrdi"),
		S3Endpoint: getEnv("S3_ENDPOINT", "http://localstack:4566"),
		AWSRegion:  getEnv("AWS_REGION", "us-east-1"),
	}
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}
