package config

import "os"

type Config struct {
	AppPort        string
	MongoURI       string
	MongoDBName    string
	RootUsername   string
	RootPassword   string
}

func NewConfig() *Config {
	return &Config{
		AppPort:        getEnv("APP_PORT", "3000"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:    getEnv("MONGO_DB_NAME", "scripts_management"),
		RootUsername:   getEnv("ROOT_USERNAME", "root"),
		RootPassword:   getEnv("ROOT_PASSWORD", "root123"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
