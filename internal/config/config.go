package config

import "os"

type Config struct {
	AppPort     string
	MongoURI    string
	MongoDBName string
}

func NewConfig() *Config {
	return &Config{
		AppPort:     getEnv("APP_PORT", "3000"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "scripts_management"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
