package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Simple helper function to read an environment or panic if it's missing
func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	panic(fmt.Sprintf("Missing required env variable %s", key))
}

// Simple helper function to read an environment or return a default value
func getOptionalEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getOptionalEnvAsInt(name string, defaultVal int) int {
	valueStr := getOptionalEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func getOptionalEnvAsDuration(name string, defaultVal int) time.Duration {
	valueStr := getOptionalEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return time.Duration(value)
	}

	return time.Duration(defaultVal)
}

func getOptionalEnvAsFloat32(name string, defaultVal float32) float32 {
	valueStr := getOptionalEnv(name, "")
	if value, err := strconv.ParseFloat(valueStr, 32); err == nil {
		return float32(value)
	}

	return defaultVal
}
