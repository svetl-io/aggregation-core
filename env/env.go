package env

// Load local env variables
import (
	"time"

	_ "github.com/joho/godotenv/autoload"
)

var SOURCE_KAFKA_BROKER = getEnv("SOURCE_KAFKA_BROKER")
var TOPICS = getEnv("TOPICS")
var SHUTDOWN_TIMEOUT_SECONDS = getOptionalEnvAsDuration("SHUTDOWN_TIMEOUT_SECONDS", 30) * time.Second
var LOG_LEVEL = getOptionalEnv("LOG_LEVEL", "INFO")
var PORT = getOptionalEnvAsInt("PORT", 8080)
