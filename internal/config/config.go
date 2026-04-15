package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Port                  string
	MongoURI              string
	MongoDB               string
	TransactionCollection string
	RedisAddr             string
	StaticAuthToken       string
	SeederNumUsers        int
	SeederNumRounds       int
	SeederBatchSize       int
	SeederConcurrent      bool
}

// LoadConfig loads the configuration from environment variables
func LoadConfig(log *zap.SugaredLogger) *Config {
	err := godotenv.Load()
	if err != nil {
		log.Warn("No .env file found, using system environment variables")
	}

	return &Config{
		Port:                  getEnv("PORT", "8080"),
		MongoURI:              getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:               getEnv("MONGO_DB", "casino_analytics"),
		TransactionCollection: getEnv("TRANSACTION_COLLECTION", "transactions"),
		RedisAddr:             getEnv("REDIS_ADDR", "localhost:6379"),
		StaticAuthToken:       getEnv("STATIC_AUTH_TOKEN", "super-secret-admin-key"),
		SeederNumUsers:        getEnvAsInt("SEEDER_NUM_USERS", 500),
		SeederNumRounds:       getEnvAsInt("SEEDER_NUM_ROUNDS", 2000000),
		SeederBatchSize:       getEnvAsInt("SEEDER_BATCH_SIZE", 10000),
		SeederConcurrent:      getEnvAsBool("SEEDER_CONCURRENT", true),
	}
}

// getEnv gets the environment variable with the given key and fallback value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvAsInt gets the environment variable with the given key and fallback value
func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

// getEnvAsBool gets the environment variable with the given key and fallback value
func getEnvAsBool(key string, fallback bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return fallback
}
